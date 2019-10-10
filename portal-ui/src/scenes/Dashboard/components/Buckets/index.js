import React from 'react';
import { compose, withState } from 'recompose'
import { connect } from 'react-redux';

import { Modal } from 'react-bootstrap';

import './styles.css';

function Buckets(props) {
  return (
    <div className="Buckets">
      <div className="row"><h1>Buckets</h1></div>
      <div className="row">
        <button type="button" className="btn btn-outline-success mr-1" onClick={() => props.setShow(true)}>Create Bucket</button>
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

      <Modal show={props.show} onHide={() => props.setShow(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Create Bucket</Modal.Title>
        </Modal.Header>
        <Modal.Body>

        <form className="col-sm-6">
          <div className="form-group">
            <label htmlFor="organization">Bucket Name</label>
            <input type="text" name="organization" placeholder="" required className="form-control" />
          </div>
          <div className="form-group">
            <label htmlFor="policy">Policy</label>
            <input type="text" name="policy" placeholder="" required className="form-control" />
          </div>
          <div className="form-group">
            <label htmlFor="groups">Groups</label>
            <input type="password" name="groups" placeholder="" required className="form-control" />
          </div>
        </form>

        </Modal.Body>
        <Modal.Footer>
          <button className="btn btn-outline-success" onClick={() => props.setShow(false)}>Close</button>
          <button className="btn btn-success" onClick={() => props.setShow(false)}>Save</button>
        </Modal.Footer>
      </Modal>

    </div>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
  withState('show', 'setShow', false)
)(Buckets);