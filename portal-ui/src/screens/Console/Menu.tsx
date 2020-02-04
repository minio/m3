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
import DashboardIcon from "@material-ui/icons/Dashboard";
import BarChartIcon from "@material-ui/icons/BarChart";
import LayersIcon from "@material-ui/icons/Layers";
import ArchiveIcon from "@material-ui/icons/Archive";
import { Link } from "react-router-dom";
import { Divider } from "@material-ui/core";
import { ExitToApp } from "@material-ui/icons";
import { AppState } from "../../store";
import { connect } from "react-redux";
import { userLoggedIn } from "../../actions";
import List from "@material-ui/core/List";
import storage from "local-storage-fallback";
import history from "../../history";
import LockIcon from "@material-ui/icons/Lock";

const mapState = (state: AppState) => ({
  open: state.system.loggedIn
});

const connector = connect(mapState, { userLoggedIn });

interface MenuProps {
  userLoggedIn: typeof userLoggedIn;
}

class Menu extends React.Component<MenuProps> {
  logout() {
    storage.removeItem("token");
    this.props.userLoggedIn(false);
    history.push("/");
  }

  render() {
    return (
      <List>
        {}
        <div>
          <ListItem button component={Link} to="/">
            <ListItemIcon>
              <DashboardIcon />
            </ListItemIcon>
            <ListItemText primary="Dashboard" />
          </ListItem>
          <ListItem button component={Link} to="/buckets">
            <ListItemIcon>
              <ArchiveIcon />
            </ListItemIcon>
            <ListItemText primary="Buckets" />
          </ListItem>
          <ListItem button component={Link} to="/permissions">
            <ListItemIcon>
              <LockIcon />
            </ListItemIcon>
            <ListItemText primary="Permissions" />
          </ListItem>
          <ListItem button>
            <ListItemIcon>
              <BarChartIcon />
            </ListItemIcon>
            <ListItemText primary="Reports" />
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
        </div>
      </List>
    );
  }
}

export default connector(Menu);
