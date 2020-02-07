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
import { createStyles, Theme, withStyles } from "@material-ui/core/styles";
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
  IconButton,
  LinearProgress,
  TableFooter,
  TablePagination
} from "@material-ui/core";
import Typography from "@material-ui/core/Typography";
import DeleteIcon from "@material-ui/icons/Delete";
import AddBucket from "./AddBucket";
import DeleteBucket from "./DeleteBucket";
import { MinTablePaginationActions } from "../../../common/MinTablePaginationActions";
import { CreateIcon } from "../../../icons";

const styles = (theme: Theme) =>
  createStyles({
    seeMore: {
      marginTop: theme.spacing(3)
    },
    paper: {
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
    },
    minTableHeader: {
      color: "#393939",
      "& tr": {
        "& th": {
          fontWeight: "bold"
        }
      }
    },
    actionsTray: {
      textAlign: "right"
    }
  });

interface IBucketsProps {
  classes: any;
}

interface IBucketsState {
  records: Bucket[];
  totalRecords: number;
  loading: boolean;
  error: string;
  deleteError: string;
  addScreenOpen: boolean;
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
    error: "",
    deleteError: "",
    addScreenOpen: false,
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
          // if we get 0 results, and page > 0 , go down 1 page
          if (
            (res.buckets === undefined ||
              res.buckets == null ||
              res.buckets.length === 0) &&
            page > 0
          ) {
            const newPage = page - 1;
            this.setState({ page: newPage }, () => {
              this.fetchRecords();
            });
          }
        })
        .catch(err => {
          this.setState({ loading: false, error: err });
        });
    });
  }

  closeAddModalAndRefresh() {
    this.setState({ addScreenOpen: false }, () => {
      this.fetchRecords();
    });
  }

  closeDeleteModalAndRefresh(refresh: boolean) {
    this.setState({ deleteOpen: false }, () => {
      if (refresh) {
        this.fetchRecords();
      }
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
      this.setState({ page: 0, rowsPerPage: rPP }, () => {
        this.fetchRecords();
      });
    };

    const confirmDeleteBucket = (bucket: string) => {
      this.setState({ deleteOpen: true, selectedBucket: bucket });
    };

    const bytesToSize = (bytesStr: string) => {
      const bytes = parseInt(bytesStr, 10);
      if (isNaN(bytes)) {
        return "0 Byte";
      }
      const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
      if (bytes === 0) return "0 Byte";
      const i = Math.floor(Math.log(bytes) / Math.log(1024));
      return Math.round(bytes / Math.pow(1024, i)) + " " + sizes[i];
    };

    return (
      <React.Fragment>
        <AddBucket
          open={addScreenOpen}
          closeModalAndRefresh={() => {
            this.closeAddModalAndRefresh();
          }}
        />
        <Grid container>
          <Grid item xs={12}>
            <Typography variant="h6">Buckets</Typography>
          </Grid>
          <Grid item xs={12}>
            <br />
          </Grid>
          <Grid item xs={6}></Grid>
          <Grid item xs={6} className={classes.actionsTray}>
            <Button
              variant="contained"
              color="primary"
              startIcon={<CreateIcon />}
              onClick={() => {
                this.setState({
                  addScreenOpen: true
                });
              }}
            >
              Create Bucket
            </Button>
          </Grid>
          <Grid item xs={12}>
            <br />
          </Grid>
          <Grid item xs={12}>
            <Paper className={classes.paper}>
              {loading && <LinearProgress />}
              {records != null && records.length > 0 ? (
                <Table size="small">
                  <TableHead className={classes.minTableHeader}>
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
                        <TableCell>{bytesToSize(row.size)}</TableCell>
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
                        rowsPerPageOptions={[5, 10, 25]}
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
                        ActionsComponent={MinTablePaginationActions}
                      />
                    </TableRow>
                  </TableFooter>
                </Table>
              ) : (
                <div>No Buckets</div>
              )}
            </Paper>
          </Grid>
        </Grid>

        <DeleteBucket
          deleteOpen={deleteOpen}
          selectedBucket={selectedBucket}
          closeDeleteModalAndRefresh={(refresh: boolean) => {
            this.closeDeleteModalAndRefresh(refresh);
          }}
        />
      </React.Fragment>
    );
  }
}

export default withStyles(styles)(Buckets);
