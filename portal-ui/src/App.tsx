import React from 'react';
import {Route, Router, Switch} from "react-router-dom";
import Landing from "./landing";
import Signup from "./signup";
import history from "./history";
import Login from "./login";


const App: React.FC = () => {
    return (
        <Router history={history}>
            <Switch>
                <Route path="/signup">
                    <Signup/>
                </Route>
                <Route path="/login">
                    <Login />
                </Route>
                <Route exact path="/">
                    <Landing/>
                </Route>
            </Switch>
        </Router>
    );
};

export default App;
