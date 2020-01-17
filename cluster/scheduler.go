// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/m3/cluster/db"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TaskStatus string

const (
	NewTaskStatus             TaskStatus = "new"
	ScheduledTaskStatus                  = "scheduled"
	CompleteTaskStatus                   = "complete"
	ErrorSchedulingTaskStatus            = "error_scheduling"
	FailedTaskStatus                     = "failed"
	StalledTaskStatus                    = "stalled"
	UnknownTaskStatus                    = "unknown"
)

const (
	failTask              = "fail"
	emptyTask             = "empty"
	TaskProvisionTenant   = "provision-tenant"
	TaskDeprovisionTenant = "deprovision-tenant"
	TaskSendEmailToUser   = "send-email-to-user"
	TaskSendAdminInvite   = "send-admin-invite"
)

type Task struct {
	ID     int64
	Name   string
	Status TaskStatus
	// json representation of the data
	Data []byte
}

// starts a loop that monitors the tasks table for pending task to schedule inside the cluster
func StartScheduler() {
	// monitor tasks table for new tasks every 500ms
	for {
		ctx, err := NewEmptyContext()
		if err != nil {
			panic(err)
		}
		task, err := fetchNewTask(ctx)
		if err != nil && err != sql.ErrNoRows {
			// panic if we can't fetch tasks
			panic(err)
		}
		// if we got a task, schedule it
		if task != nil {
			log.Printf("Schedule task %d\n", task.ID)
			if err := scheduleTaskJob(ctx, task); err != nil {
				log.Println(err)
				if err = markTask(ctx, task, ErrorSchedulingTaskStatus); err != nil {
					log.Println(err)
				}
			}
			if err := ctx.Commit(); err != nil {
				log.Println(err)
			}
		} else {
			if err := ctx.Rollback(); err != nil {
				log.Println(err)
			}
			// we got not task, sleep a little
			time.Sleep(time.Millisecond * 500)
		}
	}
}

// fetchNewTask gets a task in "new" state and locks it until it's unlocked by an update to the record.
// We do the locking at database just in case there's 2 schedulers running, they cannot grab the same task.
func fetchNewTask(ctx *Context) (*Task, error) {
	// select a task in new state and lock it
	query :=
		`SELECT 
				t.id, t.name, t.data, t.status
			FROM 
				tasks t
			WHERE t.status=$1
			LIMIT 1
		FOR UPDATE`
	// query the reord
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	row := tx.QueryRow(query, NewTaskStatus)
	task := Task{}
	// Save the resulted query on the User struct
	err = row.Scan(&task.ID, &task.Name, &task.Data, &task.Status)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// markTask marks a task as the passes `newStatus` state.
func markTask(ctx *Context, task *Task, newStatus TaskStatus) error {
	// if we are marking the task as scheduled, set the schedule time field
	extraField := ""
	if newStatus == ScheduledTaskStatus {
		extraField = ", scheduled_time = now()"
	}
	// build the query
	query := fmt.Sprintf(
		`UPDATE tasks 
					SET status = $1 %s
				WHERE id=$2`, extraField)
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Execute Query
	_, err = tx.Exec(query, newStatus, task.ID)
	if err != nil {
		return err
	}
	return nil
}

// scheduleTaskJob creates the kubernetes job for the passed task and if successful, it marks the task as scheduled
func scheduleTaskJob(ctx *Context, task *Task) error {
	// create job
	err := startJob(task)
	if err != nil {
		log.Println(err)
		if err = markTask(ctx, task, ErrorSchedulingTaskStatus); err != nil {
			log.Println(err)
		}
		return err
	}
	// mark the task as scheduled
	if err = markTask(ctx, task, ScheduledTaskStatus); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// startJob starts a kubernets job for the passed task
func startJob(task *Task) error {
	var backoff int32 = 0
	var ttlJob int32 = 60
	log.Println("Scheduling job:", fmt.Sprintf("task-%d-%s-job", task.ID, task.Name))
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("task-%d-%s-job", task.ID, task.Name),
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlJob,
			BackoffLimit:            &backoff,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: "m3-user",
					Containers: []v1.Container{
						{
							Name:  "task",
							Image: getM3ContainerImage(),
							Command: []string{
								"/m3",
								"run-task",
								fmt.Sprintf("%d", task.ID),
							},
							ImagePullPolicy: v1.PullPolicy(getM3ImagePullPolicy()),
							EnvFrom: []v1.EnvFromSource{
								{
									ConfigMapRef: &v1.ConfigMapEnvSource{
										LocalObjectReference: v1.LocalObjectReference{Name: "m3-env"},
									},
								},
							},
						},
					},
					RestartPolicy: "Never",
					// TODO: Select only application nodes
				},
			},
		},
	}
	// schedule job in k8s
	clientset, err := k8sClient()
	if err != nil {
		return err
	}
	_, err = clientset.BatchV1().Jobs(defNS).Create(&job)
	if err != nil {
		return err
	}
	return nil
}

// RunTask runs a task by id and records the result of if on the task record.
// attempts to recover from a panic in case there's one within the task and also marks it on the db.
func RunTask(id int64) error {
	task, err := getTaskByID(id)
	if err != nil {
		return err
	}
	ctx, err := NewEmptyContext()
	if err != nil {
		return err
	}
	// if the task fails, mark it as such
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			if err := markTask(ctx, task, FailedTaskStatus); err != nil {
				log.Println(err)
			}
			if err := ctx.Commit(); err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second * 2)
			os.Exit(1)
		}
		if err := ctx.Rollback(); err != nil {
			log.Println(err)
		}
	}()
	// check whether the task name is a registered task name or not
	switch task.Name {
	case failTask:
		panic("Intentional Task Failure")
	case emptyTask:
		log.Println("Empty Task")
	case TaskProvisionTenant:
		if err := ProvisionTenantTask(task); err != nil {
			panic(err)
		}
	case TaskDeprovisionTenant:
		if err := DeprovisionTenantTask(task); err != nil {
			panic(err)
		}
	case TaskSendEmailToUser:
		if err := SendEmailToUserTask(task); err != nil {
			panic(err)
		}
	case TaskSendAdminInvite:
		if err := SendAdminInviteTask(task); err != nil {
			panic(err)
		}
	default:
		log.Printf("Unknown task name: %s\n", task.Name)
		if err := markTask(ctx, task, UnknownTaskStatus); err != nil {
			log.Println(err)
		}
	}
	// mark the task as complete
	if err = markTask(ctx, task, CompleteTaskStatus); err != nil {
		return err
	}
	if err = ctx.Commit(); err != nil {
		log.Println(err)
	}
	os.Exit(0)
	return nil
}

// getTaskByID returns a task by id
func getTaskByID(id int64) (*Task, error) {
	query :=
		`SELECT 
				t.id, t.name, t.data, t.status
			FROM 
				tasks t
			WHERE t.id=$1
			LIMIT 1`
	// query the reord
	row := db.GetInstance().Db.QueryRow(query, id)
	task := Task{}
	// Save the resulted query on the User struct
	err := row.Scan(&task.ID, &task.Name, &task.Data, &task.Status)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func ScheduleTask(ctx *Context, name string, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Now insert the credentials into the DB
	query := `
		INSERT INTO
				tasks ("name", "data")
			  VALUES
				($1, $2)`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, name, string(dataJSON))
	if err != nil {
		return err
	}
	return nil

}
