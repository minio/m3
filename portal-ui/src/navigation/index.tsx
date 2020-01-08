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


import React from 'react';
import {makeStyles} from '@material-ui/core/styles';
import {Route, Router, Switch} from "react-router-dom";
import Dashboard from "../dashboard";
import history from "../history";

const useStyles = makeStyles(theme => ({
    '@global': {
        body: {
            backgroundColor: theme.palette.common.white,
        },
    },

    errorBlock: {
        color: 'red',
    }
}));
const Navigation: React.FC = () => {

    return (
        <Router history={history}>
            <Switch>
                <Route
                    path={
                        "/dashboard"
                    }
                    component={Dashboard}
                />
            </Switch>
        </Router>
    );
};

export default Navigation;
