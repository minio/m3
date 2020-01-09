import React from 'react';
import {Route, Router, Switch} from "react-router-dom";
import Landing from "./landing";
import Signup from "./signup";
import history from "./history";
import Login from "./login";
import Navigation from "./navigation";
import storage from 'local-storage-fallback';


const App: React.FC = () => {

  const isLoggedIn = () => {
    return storage.getItem('token') !== undefined &&
      storage.getItem('token') !== null &&
      storage.getItem('token') !== '';
  }
  return (
    <Router history={history}>
      <Switch>
        <Route path="/signup">
          <Signup/>
        </Route>
        <Route path="/login">
          <Login/>
        </Route>
        <Route path="/">
          {isLoggedIn() ? (<Navigation/>) : (<Landing/>)}
        </Route>
      </Switch>
    </Router>
  );
};

export default App;
