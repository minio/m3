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
import { makeStyles, useTheme } from "@material-ui/core/styles";
import clsx from "clsx";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import Typography from "@material-ui/core/Typography";
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Legend,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";

const useStyles = makeStyles(theme => ({
  root: {
    display: "flex"
  },
  toolbar: {
    paddingRight: 24 // keep right padding when drawer closed
  },
  toolbarIcon: {
    display: "flex",
    alignItems: "center",
    justifyContent: "flex-end",
    padding: "0 8px",
    ...theme.mixins.toolbar
  },
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
    transition: theme.transitions.create(["width", "margin"], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen
    })
  },

  menuButton: {
    marginRight: 36
  },
  menuButtonHidden: {
    display: "none"
  },
  title: {
    flexGrow: 1
  },
  drawerPaperClose: {
    overflowX: "hidden",
    transition: theme.transitions.create("width", {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen
    }),
    width: theme.spacing(7),
    [theme.breakpoints.up("sm")]: {
      width: theme.spacing(9)
    }
  },
  appBarSpacer: theme.mixins.toolbar,
  content: {
    flexGrow: 1,
    height: "100vh",
    overflow: "auto"
  },
  container: {
    paddingBottom: theme.spacing(4),
    "& h6": {
      color: "#777777",
      fontSize: 14,
    }
  },
  paper: {
    padding: theme.spacing(2),
    display: "flex",
    overflow: "auto",
    flexDirection: "column"
  },
  fixedHeight: {
    minHeight: 240,
  },
  consumptionValue: {
    color: "#000000",
    fontSize: "60px",
    fontWeight: "bold",
  },
}));

export default function Dashboard() {
  const theme = useTheme();
  const classes = useStyles(theme);
  const fixedHeightPaper = clsx(classes.paper, classes.fixedHeight);

  const data = [
    {
      name: 'Page A', uv: 4000, pv: 2400, amt: 2400,
    },
    {
      name: 'Page B', uv: 3000, pv: 1398, amt: 2210,
    },
    {
      name: 'Page C', uv: 2000, pv: 9800, amt: 2290,
    },
    {
      name: 'Page D', uv: 2780, pv: 3908, amt: 2000,
    },
    {
      name: 'Page E', uv: 1890, pv: 4800, amt: 2181,
    },
    {
      name: 'Page F', uv: 2390, pv: 3800, amt: 2500,
    },
    {
      name: 'Page G', uv: 3490, pv: 4300, amt: 2100,
    },
  ];

  return (
    <Grid container xs={10}>
      <Grid container xs={12} spacing={3} className={classes.container}>
        <Grid item xs={12} md={4} lg={4}>
          <Paper className={fixedHeightPaper}>
            <Typography variant="h6">All Buckets</Typography>
            <Typography className={classes.consumptionValue}>238</Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4} lg={4}>
          <Paper className={fixedHeightPaper}>
            <Typography variant="h6">Usage</Typography>
            <Typography className={classes.consumptionValue}>375TB</Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4} lg={4}>
          <Paper className={fixedHeightPaper}>
            <Typography variant="h6">Egress this Month</Typography>
            <Typography className={classes.consumptionValue}>1.5TB</Typography>
          </Paper>
        </Grid>
      </Grid>
      <Grid container xs={12} spacing={3} className={classes.container}>
        <Grid item xs={8}>
          <Paper className={fixedHeightPaper}>
            <Typography variant="h6">Daily Average Usage</Typography>
            <ResponsiveContainer width="100%" height={400}>
              <LineChart
                width={500}
                height={300}
                data={data}
                margin={{
                  top: 5, right: 30, left: 20, bottom: 5,
                }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="pv" stroke="#8884d8" activeDot={{ r: 8 }} />
                <Line type="monotone" dataKey="uv" stroke="#82ca9d" />
              </LineChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
        <Grid item xs={4}>
          <Paper className={fixedHeightPaper}>
            <Typography variant="h6">Daily Network Egress</Typography>
          </Paper>
        </Grid>
      </Grid>
    </Grid>
  );
}
