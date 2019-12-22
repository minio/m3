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
	"fmt"
	"log"
	"os"
	"time"

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
	failTask  = "fail"
	emptyTask = "empty"
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
		task, err := fetchNewTask()
		if err != nil && err != sql.ErrNoRows {
			// panic if we can't fetch tasks
			panic(err)
		}
		// if we got a task, schedule it
		if task != nil {
			log.Printf("Schedule task %d\n", task.ID)
			if err := scheduleTask(task); err != nil {
				log.Println(err)
				if err = markTask(task, ErrorSchedulingTaskStatus); err != nil {
					log.Println(err)
				}
			}
		} else {
			// we got not task, sleep a little
			time.Sleep(time.Millisecond * 500)
		}
	}
}

// fetchNewTask gets a task in "new" state and locks it until it's unlocked by an update to the record.
// We do the locking at database just in case there's 2 schedulers running, they cannot grab the same task.
func fetchNewTask() (*Task, error) {
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
	row := GetInstance().Db.QueryRow(query, NewTaskStatus)
	task := Task{}
	// Save the resulted query on the User struct
	err := row.Scan(&task.ID, &task.Name, &task.Data, &task.Status)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// markTask marks a task as the passes `newStatus` state.
func markTask(task *Task, newStatus TaskStatus) error {
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

	// Execute Query
	_, err := GetInstance().Db.Exec(query, newStatus, task.ID)
	if err != nil {
		return err
	}
	return nil
}

// scheduleTask creates the kubernetes job for the passed task and if successful, it marks the task as scheduled
func scheduleTask(task *Task) error {
	// create job
	err := startJob(task)
	if err != nil {
		log.Println(err)
		if err = markTask(task, ErrorSchedulingTaskStatus); err != nil {
			log.Println(err)
		}
		return err
	}
	// mark the task as scheduled
	if err = markTask(task, ScheduledTaskStatus); err != nil {
		return err
	}
	return nil
}

// startJob starts a kubernets job for the passed task
func startJob(task *Task) error {
	var backoff int32 = 0
	var ttlJob int32 = 60
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
					Volumes: nil,
					Containers: []v1.Container{
						{
							Name:  "task",
							Image: getM3ContainerImage(),
							Command: []string{
								"/m3",
								"run-task",
								fmt.Sprintf("%d", task.ID),
							},
							ImagePullPolicy: "IfNotPresent",
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

// getTaskById returns a task by id
func getTaskById(id int64) (*Task, error) {
	query :=
		`SELECT 
				t.id, t.name, t.data, t.status
			FROM 
				tasks t
			WHERE t.id=$1
			LIMIT 1`
	// query the reord
	row := GetInstance().Db.QueryRow(query, id)
	task := Task{}
	// Save the resulted query on the User struct
	err := row.Scan(&task.ID, &task.Name, &task.Data, &task.Status)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// RunTask runs a task by id and records the result of if on the task record.
// attempts to recover from a panic in case there's one within the task and also marks it on the db.
func RunTask(id int64) error {
	task, err := getTaskById(id)
	if err != nil {
		return err
	}
	// if the task fails, mark it as such
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			if err := markTask(task, FailedTaskStatus); err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second * 2)
			os.Exit(1)
		}
	}()
	// check whether the task name is a registered task name or not
	switch task.Name {
	case failTask:
		panic("Intentional Task Failure")
	case emptyTask:
		log.Println("Empty Task")
	default:
		log.Printf("Unknown task name: %s\n", task.Name)
		if err := markTask(task, UnknownTaskStatus); err != nil {
			log.Println(err)
		}
	}
	// mark the task as complete
	if err = markTask(task, CompleteTaskStatus); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
