import React from 'react';
import { compose, withHandlers } from 'recompose'
import { connect } from 'react-redux';

import './styles.css';

function SideBar(props) {
  return (
    <div className="SideBar">
      <ul className="list-group menu">
        <li><a onClick={() => props.dispatchSelectSection('metrics')} href="#" className="list-group-item list-group-item-action">Dashboard</a></li>
        <li><a onClick={() => props.dispatchSelectSection('buckets')} href="#" className="list-group-item list-group-item-action">Buckets</a></li>
        <li><a onClick={() => props.dispatchSelectSection('users')} href="#" className="list-group-item list-group-item-action">Users</a></li>
        <li><a onClick={() => props.dispatchSelectSection('groups')} href="#" className="list-group-item list-group-item-action">Groups</a></li>
        <li><a onClick={() => props.dispatchSelectSection('policies')} href="#" className="list-group-item list-group-item-action">Policies</a></li>
        <li><a onClick={() => props.dispatchSelectSection('billing')} href="#" className="list-group-item list-group-item-action">Billing</a></li>
      </ul>
    </div>
  );
}

const mapStateToProps = state => ({

});

export default compose(
  connect(mapStateToProps),
  withHandlers({
    dispatchSelectSection: ({ dispatch, sectionActions }) => section => {
      dispatch(sectionActions.select(section));
    },
  })
)(SideBar);
