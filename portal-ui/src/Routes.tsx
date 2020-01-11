// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2020 MinIO, Inc.
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
import {Router, Switch, Route, Redirect, RouteProps} from "react-router-dom";
import history from "./history";
import Login from "./screens/LoginPage";
import Signup from "./screens/SignupPage";
import LandingPage from "./screens/LandingPage";
import Dashboard from './screens/Dashboard';
import NotFoundPage from './screens/NotFoundPage'
import storage from "local-storage-fallback";

const isLoggedIn = () => {
  return storage.getItem('token') !== undefined &&
    storage.getItem('token') !== null &&
    storage.getItem('token') !== '';
}

const Routes: React.FC = () => (
  <Router history={history}>
    <Switch>
      <Route exact path="/login" component={Login}/>
      <Route exact path="/signup" component={Signup}/>
      {
        isLoggedIn() ? (
          <Switch>
            <Route exact path="/dashboard" component={Dashboard}/>
            <Redirect exact from="/" to="dashboard"/>
            <Route component={NotFoundPage} />
          </Switch>
        ) : (
          <Switch>
            <Route exact path="/" component={LandingPage}/>
            <Route component={NotFoundPage} />
          </Switch>
        )
      }
    </Switch>
  </Router>
);

export default Routes;
