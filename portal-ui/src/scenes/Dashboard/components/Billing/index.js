import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';
import { Bar, Line } from 'react-chartjs-2';

import './styles.css';

function Billing(props) {
  return (
    <div className="Metrics">
        <div className="row"><h1>Billing</h1></div>
        <div className="row">
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <h1>372 TB</h1>
              <p>Capacity</p>
            </div>
          </div>          
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <h1>1.5 TB</h1>
              <p>Egress this Month</p>
            </div>
          </div>          
          <div className="col-sm-4">              
            <div className="metric">
              <h1>$238.50</h1>
              <p>Charges so for this month</p>
            </div> 
          </div>          
        </div>
        <div className="row">
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <Line data={{
                  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                  datasets: [
                    {
                      label: 'Usage VS Last Month',
                      fill: false,
                      lineTension: 0.1,
                      backgroundColor: 'rgba(75,192,192,0.4)',
                      borderColor: 'rgba(75,192,192,1)',
                      borderCapStyle: 'butt',
                      borderDash: [],
                      borderDashOffset: 0.0,
                      borderJoinStyle: 'miter',
                      pointBorderColor: 'rgba(75,192,192,1)',
                      pointBackgroundColor: '#fff',
                      pointBorderWidth: 1,
                      pointHoverRadius: 5,
                      pointHoverBackgroundColor: 'rgba(75,192,192,1)',
                      pointHoverBorderColor: 'rgba(220,220,220,1)',
                      pointHoverBorderWidth: 2,
                      pointRadius: 1,
                      pointHitRadius: 10,
                      data: [65, 79, 90, 91, 116, 145, 160]
                    }
                  ]
                }} />
            </div>
          </div>          
          <div className="col-sm-4">
            <div className="metric">
              <Line data={{
                labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                datasets: [
                  {
                    label: 'Egress VS Last Month',
                    fill: false,
                    lineTension: 0.1,
                    backgroundColor: 'rgba(75,192,192,0.4)',
                    borderColor: 'rgba(75,192,192,1)',
                    borderCapStyle: 'butt',
                    borderDash: [],
                    borderDashOffset: 0.0,
                    borderJoinStyle: 'miter',
                    pointBorderColor: 'rgba(75,192,192,1)',
                    pointBackgroundColor: '#fff',
                    pointBorderWidth: 1,
                    pointHoverRadius: 5,
                    pointHoverBackgroundColor: 'rgba(75,192,192,1)',
                    pointHoverBorderColor: 'rgba(220,220,220,1)',
                    pointHoverBorderWidth: 2,
                    pointRadius: 1,
                    pointHitRadius: 10,
                    data: [32, 48, 45, 57, 83, 102, 112]
                  }
                ]
              }} />
            </div> 
          </div>          
          <div className="col-sm-4">              
            <div className="metric">
              <Line data={{
                labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                datasets: [
                  {
                    label: 'Charges vs Last Month',
                    fill: false,
                    lineTension: 0.1,
                    backgroundColor: 'rgba(75,192,192,0.4)',
                    borderColor: 'rgba(75,192,192,1)',
                    borderCapStyle: 'butt',
                    borderDash: [],
                    borderDashOffset: 0.0,
                    borderJoinStyle: 'miter',
                    pointBorderColor: 'rgba(75,192,192,1)',
                    pointBackgroundColor: '#fff',
                    pointBorderWidth: 1,
                    pointHoverRadius: 5,
                    pointHoverBackgroundColor: 'rgba(75,192,192,1)',
                    pointHoverBorderColor: 'rgba(220,220,220,1)',
                    pointHoverBorderWidth: 2,
                    pointRadius: 1,
                    pointHitRadius: 10,
                    data: [65, 59, 80, 81, 56, 55, 40]
                  }
                  ]
                }}
              />
            </div>          
        </div>
        </div>
        <div className="row">
          <h1>Statements</h1>
          <table className="table table-bordered">
            <thead>
              <tr>
                <th width="10%" scope="col">Select</th>
                <th scope="col">Month^y</th>
                <th scope="col">Total</th>
                <th scope="col">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <th scope="row"><input type="checkbox" checked/></th>
                <td>September 2019</td>
                <td>$358.45</td>
                <td><button className="btn btn-success">Download</button></td>
              </tr>
              <tr>
                <th scope="row"><input type="checkbox"/></th>
                <td>August 2019</td>
                <td>$353.49</td>
                <td><button className="btn btn-success">Download</button></td>
              </tr>
              <tr>
                <th scope="row"><input type="checkbox"/></th>
                <td>July 2019</td>
                <td>$351.25</td>
                <td><button className="btn btn-success">Download</button></td>
              </tr>
              <tr>
                <th scope="row"><input type="checkbox"/></th>
                <td>June 2019</td>
                <td>$344.44</td>
                <td><button className="btn btn-success">Download</button></td>
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
)(Billing);