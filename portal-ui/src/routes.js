import React from 'react';
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";
import LandingPage from './scenes/LandingPage';
import SignUp from './scenes/SignUp';

const Routes = (props) => (
 <Router {...props}>
    <Switch>
      <Route exact path="/">
        <LandingPage />
      </Route>
      <Route path="/home">
        <LandingPage />
      </Route>
      <Route path="/pricing">
        <LandingPage />
      </Route>
      <Route path="/about-us">
        <LandingPage />
      </Route>
      <Route path="/signup">
        <SignUp />
      </Route>
    </Switch>
 </Router>
);
export default Routes;