import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';

import './styles.css';

function Buckets(props) {
  return (
    <div className="Buckets">
      <div className="row"><h1>Buckets</h1></div>
      <div className="row">
        <button type="button" className="btn btn-outline-success mr-1">Create Bucket</button>
        <button type="button" className="btn btn-outline-success">Assign Policy</button>
      </div>
      <div className="row top-buffer">
      <table className="table table-bordered">
  <thead>
    <tr>
      <th width="10%" scope="col">Select</th>
      <th scope="col">Name</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row"><input type="checkbox" checked/></th>
      <td>Research</td>
    </tr>
    <tr>
      <th scope="row"><input type="checkbox"/></th>
      <td>Testing</td>
    </tr>
    <tr>
      <th scope="row"><input type="checkbox"/></th>
      <td>Storage</td>
    </tr>
    <tr>
      <th scope="row"><input type="checkbox"/></th>
      <td>Production1</td>
    </tr>
    <tr>
      <th scope="row"><input type="checkbox"/></th>
      <td>Production2</td>
    </tr>
    <tr>
      <th scope="row"><input type="checkbox"/></th>
      <td>Production3</td>
    </tr>
  </tbody>
</table>
      </div>
    </div>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
)(Buckets);