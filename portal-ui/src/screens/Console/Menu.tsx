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
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import LayersIcon from "@material-ui/icons/Layers";
import { Link } from "react-router-dom";
import { Divider, withStyles } from "@material-ui/core";
import { ExitToApp } from "@material-ui/icons";
import { AppState } from "../../store";
import { connect } from "react-redux";
import { userLoggedIn } from "../../actions";
import List from "@material-ui/core/List";
import storage from "local-storage-fallback";
import history from "../../history";
import logo from "../../icons/mkube_logo_temp.svg";
import {
  BucketsIcon,
  DashboardIcon,
  PermissionIcon,
  ServiceAccountIcon
} from "../../icons";
import { createStyles, Theme } from "@material-ui/core/styles";

const styles = (theme: Theme) =>
  createStyles({
    logo: {
      paddingTop: "42px",
      marginBottom: "20px",
      textAlign: "center",
      "& img": {
        width: "76px"
      }
    },
    menuList: {
      paddingLeft: "30px",
      "& .MuiSvgIcon-root": {
        fontSize: "18px",
        color: "#393939"
      },
      "& .MuiListItemIcon-root": {
        minWidth: "40px"
      },
      "& .MuiTypography-root": {
        fontSize: "14px"
      },
      "& .MuiListItem-gutters": {
        paddingRight: "0px"
      }
    }
  });

const mapState = (state: AppState) => ({
  open: state.system.loggedIn
});

const connector = connect(mapState, { userLoggedIn });

interface MenuProps {
  classes: any;
  userLoggedIn: typeof userLoggedIn;
}

class Menu extends React.Component<MenuProps> {
  logout() {
    storage.removeItem("token");
    this.props.userLoggedIn(false);
    history.push("/");
  }

  render() {
    const { classes } = this.props;
    return (
      <React.Fragment>
        <div className={classes.logo}>
          <img src={logo} />
        </div>
        <List className={classes.menuList}>
          <ListItem button component={Link} to="/">
            <ListItemIcon>
              <DashboardIcon />
            </ListItemIcon>
            <ListItemText primary="Dashboard" />
          </ListItem>
          <ListItem button component={Link} to="/buckets">
            <ListItemIcon>
              <BucketsIcon />
            </ListItemIcon>
            <ListItemText primary="Buckets" />
          </ListItem>
          <ListItem button component={Link} to="/permissions">
            <ListItemIcon>
              <PermissionIcon />
            </ListItemIcon>
            <ListItemText primary="Permissions" />
          </ListItem>
          <ListItem button component={Link} to="/service_accounts">
            <ListItemIcon>
              <ServiceAccountIcon />
            </ListItemIcon>
            <ListItemText primary="Service Accounts" />
          </ListItem>
          <ListItem button>
            <ListItemIcon>
              <LayersIcon />
            </ListItemIcon>
            <ListItemText primary="Integrations" />
          </ListItem>
          <Divider />
          <ListItem
            button
            onClick={() => {
              this.logout();
            }}
          >
            <ListItemIcon>
              <ExitToApp />
            </ListItemIcon>
            <ListItemText primary="Logout" />
          </ListItem>
        </List>
      </React.Fragment>
    );
  }
}

export default connector(withStyles(styles)(Menu));
