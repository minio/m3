import React from 'react';
import { compose, withHandlers, lifecycle } from 'recompose'
import { connect } from 'react-redux';

function NavigationBar(props) {
  return (
    <nav className="navbar navbar-expand-lg navbar-light bg-light">
      <div className="collapse navbar-collapse" id="navbarSupportedContent">
        <ul className="navbar-nav mr-auto">
          <li className="nav-item active">
            <a className="nav-link" href="#">Home</a>
          </li>
          <li className="nav-item">
            <a className="nav-link" href="#">Pricing</a>
          </li>
          <li className="nav-item">
            <a className="nav-link" href="#">About Us</a>
          </li>
          <li className="nav-item">
            <a className="btn btn-outline-success my-2 my-sm-0" href="#">Sign Up</a>
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