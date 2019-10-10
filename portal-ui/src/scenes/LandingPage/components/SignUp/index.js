import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';
import { Link } from "react-router-dom";

import './styles.css';

function SignUp(props) {
  return (
    <div className="SignUp">
      <div className="container">
        <div className="row justify-content-center">
        <form>
          <div className="form-group">
            <label htmlFor="username">Full Name</label>
            <input type="text" name="username" placeholder="Enter your full name" required className="form-control" />
          </div>
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
          <div className="form-group">
            <label htmlFor="cardNumber">Card number</label>
            <div className="input-group">
              <input type="text" name="cardNumber" placeholder="Your card number" className="form-control" required />
              <div className="input-group-append">
                <span className="input-group-text text-muted">
                  <i className="fa fa-cc-visa mx-1"></i>
                  <i className="fa fa-cc-amex mx-1"></i>
                  <i className="fa fa-cc-mastercard mx-1"></i>
                </span>
              </div>
            </div>
          </div>
          <div className="row">
            <div className="col-sm-8">
              <div className="form-group">
                <label><span className="hidden-xs">Expiration</span></label>
                <div className="input-group">
                  <input type="number" placeholder="MM" name="" className="form-control" required />
                  <input type="number" placeholder="YY" name="" className="form-control" required />
                </div>
              </div>
            </div>
            <div className="col-sm-4">
              <div className="form-group mb-4">
                <label data-toggle="tooltip" title="Three-digits code on the back of your card">CVV
                  <i className="fa fa-question-circle"></i>
                </label>
                <input type="text" required className="form-control" />
              </div>
            </div>
          </div>
          {/* <button type="button" className="subscribe btn btn-outline-success btn-block ro÷unded-pill shadow-sm">Sign Up</button> */}
          <Link to="/dashboard" className="subscribe btn btn-outline-success btn-block rounded-pill shadow-sm">Sign Up</Link>
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
)(SignUp);