import React from 'react';
import { compose, withHandlers, lifecycle } from 'recompose'
import { connect } from 'react-redux';
import { Link } from "react-router-dom";

function NavigationBar(props) {
  return (
    <nav className="navbar navbar-expand-lg navbar-light bg-light">
      <div className="collapse navbar-collapse" id="navbarSupportedContent">
        <ul className="navbar-nav mr-auto">
          <li className="nav-item active">
            <Link className="nav-link" to="/">Home</Link>
          </li>
          <li className="nav-item">
            <Link className="nav-link" to="/pricing">Pricing</Link>
          </li>
          <li className="nav-item">
            <Link className="nav-link" to="/about-us">About Us</Link>
          </li>
          <li className="nav-item">
            <Link className="btn btn-outline-success my-2 my-sm-0" to="/signup">Sign Up</Link>
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