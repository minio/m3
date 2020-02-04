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

import React from "react";
import {
  createStyles,
  makeStyles,
  Theme,
  useTheme,
  withStyles
} from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import Paper from "@material-ui/core/Paper";
import Grid from "@material-ui/core/Grid";
import api from "../../../common/api";
import { Bucket, BucketList } from "./types";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Drawer,
  IconButton,
  LinearProgress,
  TableFooter,
  TablePagination,
  TextField,
  Toolbar
} from "@material-ui/core";
import Title from "../../../common/Title";
import Typography from "@material-ui/core/Typography";
import DeleteIcon from "@material-ui/icons/Delete";
import { KeyboardArrowLeft, KeyboardArrowRight } from "@material-ui/icons";
import { TablePaginationActionsProps } from "@material-ui/core/TablePagination/TablePaginationActions";
import FirstPageIcon from "@material-ui/icons/FirstPage";
import LastPageIcon from "@material-ui/icons/LastPage";

const styles = (theme: Theme) =>
  createStyles({
    seeMore: {
      marginTop: theme.spacing(3)
    },
    paper: {
      // padding: theme.spacing(2),
      display: "flex",
      overflow: "auto",
      flexDirection: "column"
    },
    addSideBar: {
      width: "320px",
      padding: "20px"
    },
    errorBlock: {
      color: "red"
    },
    tableToolbar: {
      paddingLeft: theme.spacing(2),
      paddingRight: theme.spacing(0)
    }
  });

const useStyles1 = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      flexShrink: 0,
      marginLeft: theme.spacing(2.5)
    }
  })
);

function TablePaginationActions(props: TablePaginationActionsProps) {
  const classes = useStyles1();
  const theme = useTheme();
  const { count, page, rowsPerPage, onChangePage } = props;

  const handleFirstPageButtonClick = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    onChangePage(event, 0);
  };

  const handleBackButtonClick = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    onChangePage(event, page - 1);
  };

  const handleNextButtonClick = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    onChangePage(event, page + 1);
  };

  const handleLastPageButtonClick = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    onChangePage(event, Math.max(0, Math.ceil(count / rowsPerPage) - 1));
  };

  return (
    <div className={classes.root}>
      <IconButton
        onClick={handleFirstPageButtonClick}
        disabled={page === 0}
        aria-label="first page"
      >
        {theme.direction === "rtl" ? <LastPageIcon /> : <FirstPageIcon />}
      </IconButton>
      <IconButton
        onClick={handleBackButtonClick}
        disabled={page === 0}
        aria-label="previous page"
      >
        {theme.direction === "rtl" ? (
          <KeyboardArrowRight />
        ) : (
          <KeyboardArrowLeft />
        )}
      </IconButton>
      <IconButton
        onClick={handleNextButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="next page"
      >
        {theme.direction === "rtl" ? (
          <KeyboardArrowLeft />
        ) : (
          <KeyboardArrowRight />
        )}
      </IconButton>
      <IconButton
        onClick={handleLastPageButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="last page"
      >
        {theme.direction === "rtl" ? <FirstPageIcon /> : <LastPageIcon />}
      </IconButton>
    </div>
  );
}

interface IBucketsProps {
  classes: any;
}

interface IBucketsState {
  records: Bucket[];
  totalRecords: number;
  loading: boolean;
  addLoading: boolean;
  deleteLoading: boolean;
  error: string;
  addError: string;
  deleteError: string;
  addScreenOpen: boolean;
  bucketName: string;
  page: number;
  rowsPerPage: number;
  deleteOpen: boolean;
  selectedBucket: string;
}

class Buckets extends React.Component<IBucketsProps, IBucketsState> {
  state: IBucketsState = {
    records: [],
    totalRecords: 0,
    loading: false,
    addLoading: false,
    deleteLoading: false,
    error: "",
    addError: "",
    deleteError: "",
    addScreenOpen: false,
    bucketName: "",
    page: 0,
    rowsPerPage: 10,
    deleteOpen: false,
    selectedBucket: ""
  };

  fetchRecords() {
    this.setState({ loading: true }, () => {
      const { page, rowsPerPage } = this.state;
      const offset = page * rowsPerPage;
      api
        .invoke("GET", `/api/v1/buckets?offset=${offset}&limit=${rowsPerPage}`)
        .then((res: BucketList) => {
          this.setState({
            loading: false,
            records: res.buckets,
            totalRecords: res.total_buckets,
            error: ""
          });
        })
        .catch(err => {
          this.setState({ loading: false, error: err });
        });
    });
  }

  addRecord(event: React.FormEvent) {
    event.preventDefault();
    const { bucketName, addLoading } = this.state;
    if (addLoading) {
      return;
    }
    this.setState({ addLoading: true }, () => {
      api
        .invoke("POST", "/api/v1/buckets", {
          name: bucketName
        })
        .then((res: BucketList) => {
          this.setState(
            {
              addLoading: false,
              records: res.buckets,
              addError: "",
              addScreenOpen: false
            },
            () => {
              this.fetchRecords();
            }
          );
        })
        .catch(err => {
          this.setState({
            addLoading: false,
            addError: err
          });
        });
    });
  }

  removeRecord() {
    const { selectedBucket, deleteLoading } = this.state;
    if (deleteLoading) {
      return;
    }
    this.setState({ deleteLoading: true }, () => {
      api
        .invoke("DELETE", `/api/v1/buckets/${selectedBucket}`, {
          name: selectedBucket
        })
        .then((res: BucketList) => {
          this.setState(
            {
              deleteLoading: false,
              records: res.buckets,
              deleteError: "",
              deleteOpen: false
            },
            () => {
              this.fetchRecords();
            }
          );
        })
        .catch(err => {
          this.setState({
            deleteLoading: false,
            deleteError: err
          });
        });
    });
  }

  componentDidMount(): void {
    this.fetchRecords();
  }

  render() {
    const { classes } = this.props;
    const {
      records,
      totalRecords,
      addScreenOpen,
      loading,
      addLoading,
      deleteLoading,
      addError,
      page,
      rowsPerPage,
      deleteOpen,
      selectedBucket
    } = this.state;

    const handleChangePage = (event: unknown, newPage: number) => {
      this.setState({ page: newPage }, () => {
        this.fetchRecords();
      });
    };

    const handleChangeRowsPerPage = (
      event: React.ChangeEvent<HTMLInputElement>
    ) => {
      const rPP = parseInt(event.target.value, 10);
      console.log(rPP);
      this.setState({ page: 0, rowsPerPage: rPP }, () => {
        this.fetchRecords();
      });
    };

    const confirmDeleteBucket = (bucket: string) => {
      this.setState({ deleteOpen: true, selectedBucket: bucket });
    };

    return (
      <React.Fragment>
        <Drawer
          anchor="right"
          open={addScreenOpen}
          onClose={() => {
            this.setState({ addScreenOpen: false });
          }}
        >
          <div className={classes.addSideBar}>
            <form
              noValidate
              autoComplete="off"
              onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
                this.addRecord(e);
              }}
            >
              <Grid container>
                <Grid item xs={12}>
                  <Title>Add Buckets</Title>
                </Grid>
                {addError !== "" && (
                  <Grid item xs={12}>
                    <Typography
                      component="p"
                      variant="body1"
                      className={classes.errorBlock}
                    >
                      {`${addError}`}
                    </Typography>
                  </Grid>
                )}
                <Grid item xs={12}>
                  <TextField
                    id="standard-basic"
                    fullWidth
                    label="Bucket Name"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      this.setState({ bucketName: e.target.value });
                    }}
                  />
                </Grid>
                <Grid item xs={12}>
                  <br />
                </Grid>
                <Grid item xs={12}>
                  <Button
                    type="submit"
                    variant="contained"
                    color="primary"
                    fullWidth
                    disabled={addLoading}
                  >
                    Save
                  </Button>
                </Grid>
                {addLoading && (
                  <Grid item xs={12}>
                    <LinearProgress />
                  </Grid>
                )}
              </Grid>
            </form>
          </div>
        </Drawer>

        <Paper className={classes.paper}>
          <Toolbar className={classes.tableToolbar}>
            <Grid justify="space-between" container>
              <Grid item xs={10}>
                <Typography
                  className={classes.title}
                  variant="h6"
                  id="tableTitle"
                >
                  Buckets
                </Typography>
              </Grid>
              <Grid item xs={2}>
                <Button
                  variant="contained"
                  color="primary"
                  onClick={() => {
                    this.setState({ addScreenOpen: true });
                  }}
                >
                  Add Bucket
                </Button>
              </Grid>
            </Grid>
          </Toolbar>
          {loading && <LinearProgress />}
          {records != null && records.length > 0 ? (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Size</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {records.map(row => (
                  <TableRow key={row.name}>
                    <TableCell>{row.name}</TableCell>
                    <TableCell>{row.size}</TableCell>
                    <TableCell align="right">
                      <IconButton
                        aria-label="delete"
                        onClick={() => {
                          confirmDeleteBucket(row.name);
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
              <TableFooter>
                <TableRow>
                  <TablePagination
                    rowsPerPageOptions={[
                      5,
                      10,
                      25,
                      { label: "All", value: -1 }
                    ]}
                    colSpan={3}
                    count={totalRecords}
                    rowsPerPage={rowsPerPage}
                    page={page}
                    SelectProps={{
                      inputProps: { "aria-label": "rows per page" },
                      native: true
                    }}
                    onChangePage={handleChangePage}
                    onChangeRowsPerPage={handleChangeRowsPerPage}
                    ActionsComponent={TablePaginationActions}
                  />
                </TableRow>
              </TableFooter>
            </Table>
          ) : (
            <div>No Buckets</div>
          )}
        </Paper>
        <Dialog
          open={deleteOpen}
          onClose={() => {
            this.setState({ deleteOpen: false });
          }}
          aria-labelledby="alert-dialog-title"
          aria-describedby="alert-dialog-description"
        >
          <DialogTitle id="alert-dialog-title">Delete Bucket</DialogTitle>
          <DialogContent>
            {deleteLoading && <LinearProgress />}
            <DialogContentText id="alert-dialog-description">
              Are you sure you want to delete bucket <b>{selectedBucket}</b>?{" "}
              <br />A bucket can only be deleted if it's empty.
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button
              onClick={() => {
                this.setState({ deleteOpen: false });
              }}
              color="primary"
              disabled={deleteLoading}
            >
              Cancel
            </Button>
            <Button
              onClick={() => {
                this.removeRecord();
              }}
              color="secondary"
              autoFocus
            >
              Delete
            </Button>
          </DialogActions>
        </Dialog>
      </React.Fragment>
    );
  }
}

export default withStyles(styles)(Buckets);
