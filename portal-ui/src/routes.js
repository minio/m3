import React from 'react';
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";
import LandingPage from './scenes/LandingPage';
import Dashboard from './scenes/Dashboard';

const Routes = (props) => (
 <Router {...props}>
    <Switch>
      <Route exact path="/">
        <LandingPage page="home" />
      </Route>
      <Route path="/pricing">
        <LandingPage page="pricing" />
      </Route>
      <Route path="/about-us">
        <LandingPage page="about-us" />
      </Route>
      <Route path="/signup">
        <LandingPage page="signup" />
      </Route>
      <Route exact path="/dashboard">
        <Dashboard />
      </Route>
    </Switch>
 </Router>
);
export default Routes;