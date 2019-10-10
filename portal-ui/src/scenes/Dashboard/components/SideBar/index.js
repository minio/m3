import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';

import './styles.css';

function SideBar(props) {
  return (
    <div className="SideBar">
      <ul className="list-group">
        <li><a href="#" className="list-group-item list-group-item-action">Dashboard</a></li>
        <li><a href="#" className="list-group-item list-group-item-action">Buckets</a></li>
        <li><a href="#" className="list-group-item list-group-item-action">Users</a></li>
        <li><a href="#" className="list-group-item list-group-item-action">Groups</a></li>
        <li><a href="#" className="list-group-item list-group-item-action">Policies</a></li>
        <li><a href="#" className="list-group-item list-group-item-action">Billing</a></li>
      </ul>
    </div>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
)(SideBar);