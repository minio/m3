import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';

import NavigationBar from './components/NavigationBar';
import SideBar from './components/SideBar';
import Metrics from './components/Metrics';
import Buckets from './components/Buckets';
import Users from './components/Users';
import Billing from './components/Billing';

import { getSelectedSection } from './selectors'
import * as sectionActionCreators from './actions';

import './styles.css';


function Dashboard(props) {
  const sections = {
    'metrics': <Metrics />,
    'buckets': <Buckets />,
    'users': <Users />,
    'billing': <Billing />,
  }
  return (
    <div className="Dashboard">
      <NavigationBar />
      <div className="container-fluid">
        <div className="row">
          <div className="col-sm-2">
            <SideBar sectionActions={sectionActionCreators} />
          </div>
          <div className="col-sm-10">
            { sections[props.selected] || <Metrics /> }
          </div>
        </div>
      </div>
    </div>
  );
}

const mapStateToProps = state => ({
  selected: getSelectedSection(state),
});

export default compose(
  connect(mapStateToProps),
)(Dashboard);