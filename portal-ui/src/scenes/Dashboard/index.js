import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';
import NavigationBar from './components/NavigationBar';

import './styles.css';

function Dashboard(props) {
  return (
    <div className="Dashboard">
      <NavigationBar />
    </div>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
)(Dashboard);