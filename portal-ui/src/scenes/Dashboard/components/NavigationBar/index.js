import React from 'react';
import { compose, withHandlers, lifecycle } from 'recompose'
import { connect } from 'react-redux';
import { Link } from "react-router-dom";

function NavigationBar(props) {
  return (
    <nav className="navbar navbar-expand-lg navbar-light bg-light">
      <a className="navbar-brand" href="#">ACME Cloud Storage</a>
      <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarText" aria-expanded="false" aria-label="Toggle navigation">
        <span className="navbar-toggler-icon"></span>
      </button>
      <div className="navbar-collapse collapse w-100 order-3 dual-collapse2" id="navbarSupportedContent">
        <ul className="navbar-nav ml-auto">
        <li className="nav-item active">
            <Link className="nav-link" to="/dashboard">Dashboard</Link>
          </li>
          <li className="nav-item">
            <Link className="nav-link" to="/account">Account</Link>
          </li>
          <li className="nav-item">
            <Link className="nav-link" to="/">Logout</Link>
          </li>
        </ul>
      </div>
    </nav>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
  withHandlers({
  }),
  lifecycle({
  }),
)(NavigationBar);