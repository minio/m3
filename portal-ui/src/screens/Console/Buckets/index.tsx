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
import Link from "@material-ui/core/Link";
import { createStyles, Theme, withStyles } from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import Title from "../Title";
import Paper from "@material-ui/core/Paper";
import Grid from "@material-ui/core/Grid";
import api from "../../../common/api";
import { Bucket, BucketList } from "./types";

function preventDefault(event: React.MouseEvent) {
  event.preventDefault();
}

const styles = (theme: Theme) =>
  createStyles({
    seeMore: {
      marginTop: theme.spacing(3)
    },
    paper: {
      padding: theme.spacing(2),
      display: "flex",
      overflow: "auto",
      flexDirection: "column"
    }
  });

interface IBucketsProps {
  classes: any;
}

interface IBucketsState {
  records: Bucket[];
  loading: boolean;
  error: string;
}

class Buckets extends React.Component<IBucketsProps, IBucketsState> {
  state: IBucketsState = {
    records: [],
    loading: false,
    error: ""
  };

  fetchRecords() {
    this.setState({ loading: true });
    api
      .invoke("GET", "/api/v1/buckets")
      .then((res: BucketList) => {
        this.setState({ loading: false, records: res.buckets });
      })
      .catch(err => {
        this.setState({ loading: false, error: err });
      });
  }

  componentDidMount(): void {
    this.fetchRecords();
  }

  render() {
    const { classes } = this.props;
    const { records } = this.state;

    return (
      <React.Fragment>
        <Grid item xs={12}>
          <Paper className={classes.paper}>
            <Title>Buckets</Title>
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
                      <TableCell align="right">Delete</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <div>No Buckets</div>
            )}
            <div className={classes.seeMore}>
              <Link color="primary" href="#" onClick={preventDefault}>
                Next
              </Link>
            </div>
          </Paper>
        </Grid>
      </React.Fragment>
    );
  }
}

export default withStyles(styles)(Buckets);
