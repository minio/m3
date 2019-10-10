import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';
import { Link } from "react-router-dom";

import './styles.css';

function Login(props) {
  return (
    <div className="Login">
      <div className="container">
        <div className="row justify-content-center">
        <form className="col-sm-6">
          <div className="form-group">
            <label htmlFor="organization">Organization</label>
            <input type="text" name="organization" placeholder="Enter your organization name" required className="form-control" />
          </div>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input type="text" name="email" placeholder="Enter your email address" required className="form-control" />
          </div>
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input type="password" name="password" placeholder="Enter your password" required className="form-control" />
          </div>
          <Link to="/dashboard" className="subscribe btn btn-outline-success btn-block rounded-pill shadow-sm">Login</Link>
        </form>
        </div>
      </div>
    </div>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
)(Login);