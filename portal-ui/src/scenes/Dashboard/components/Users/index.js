import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';

import './styles.css';

function Users(props) {
  return (
    <div className="Users">
      <div className="row"><h1>Users - Local IDP</h1></div>
      <div className="row">
        <button type="button" className="btn btn-outline-success mr-1">Create User</button>
        <button type="button" className="btn btn-outline-success mr-1">Assign Policy</button>
        <button type="button" className="btn btn-outline-success">Add to Group</button>
      </div>
      <div className="row top-buffer">
        <table className="table table-bordered">
          <thead>
            <tr>
              <th width="10%" scope="col">Select</th>
              <th scope="col">Name</th>
              <th scope="col">Email</th>
              <th scope="col">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <th scope="row"><input type="checkbox" checked/></th>
              <td>Lenin Alevski</td>
              <td>user1@nasa.gov</td>
              <td><button className="btn btn-success">Tokens</button></td>
            </tr>
            <tr>
              <th scope="row"><input type="checkbox"/></th>
              <td>Daniel Valdivia</td>
              <td>user2@nasa.gov</td>
              <td><button className="btn btn-success">Tokens</button></td>
            </tr>
            <tr>
              <th scope="row"><input type="checkbox"/></th>
              <td>Cesar Nieto</td>
              <td>user3@nasa.gov</td>
              <td><button className="btn btn-success">Tokens</button></td>
            </tr>
            <tr>
              <th scope="row"><input type="checkbox"/></th>
              <td>Anand B</td>
              <td>user4@nasa.gov</td>
              <td><button className="btn btn-success">Tokens</button></td>
            </tr>
            <tr>
              <th scope="row"><input type="checkbox"/></th>
              <td>Garima K</td>
              <td>user5@nasa.gov</td>
              <td><button className="btn btn-success">Tokens</button></td>
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
)(Users);