import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';
import { Bar, Line } from 'react-chartjs-2';

import './styles.css';

function Metrics(props) {
  return (
    <div className="Metrics">
        <div className="row"><h1>Dashboard</h1></div>
        <div className="row">
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <h1>238</h1>
              <p>Buckets</p>
            </div>
          </div>          
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <h1>375 TB</h1>
              <p>Capacity</p>
            </div>
          </div>          
          <div className="col-sm-4">              
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
                    data: [65, 59, 80, 81, 56, 55, 40]
                  }
                ]
              }} />
            </div> 
          </div>          
        </div>
        <div className="row">
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <h1>1.5 TB</h1>
              <p>Egress this Month</p>
            </div>
          </div>          
          <div className="col-sm-4">
            <div className="metric">
              <Line data={{
                labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                datasets: [
                  {
                    label: 'Network',
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
                    data: [65, 59, 80, 81, 56, 55, 40]
                  }
                  ]
                }}
              />
            </div>          
        </div>
        </div>
        <div className="row">
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <h1>20 TB</h1>
              <p>Ingress this month</p>
            </div>
          </div>          
          <div className="col-sm-4 vertical-center">
            <div className="metric">
              <Bar
                data={{
                  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                  datasets: [
                    {
                      label: 'Year of Date',
                      backgroundColor: 'rgba(255,99,132,0.2)',
                      borderColor: 'rgba(255,99,132,1)',
                      borderWidth: 1,
                      hoverBackgroundColor: 'rgba(255,99,132,0.4)',
                      hoverBorderColor: 'rgba(255,99,132,1)',
                      data: [65, 59, 80, 81, 56, 55, 40]
                    }
                  ]
                }}
                width={100}
                height={50}
              />
            </div>
          </div>          
          <div className="col-sm-4">              
            <div className="metric">
              <Bar
                data={{
                  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                  datasets: [
                    {
                      label: 'Month on Month',
                      backgroundColor: 'rgba(255,99,132,0.2)',
                      borderColor: 'rgba(255,99,132,1)',
                      borderWidth: 1,
                      hoverBackgroundColor: 'rgba(255,99,132,0.4)',
                      hoverBorderColor: 'rgba(255,99,132,1)',
                      data: [65, 59, 80, 81, 56, 55, 40]
                    }
                  ]
                }}
                width={100}
                height={50}
              />
            </div> 
          </div>          
        </div>
    </div>
  );
}

const mapStateToProps = state => ({
});

export default compose(
  connect(mapStateToProps),
)(Metrics);