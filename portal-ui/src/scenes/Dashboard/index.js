import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';
import { Bar, Line } from 'react-chartjs-2';

import NavigationBar from './components/NavigationBar';
import SideBar from './components/SideBar';

import './styles.css';

function Dashboard(props) {
  return (
    <div className="Dashboard">
      <NavigationBar />
      <div className="row">
        <div className="col-sm-2">
          <SideBar />
        </div>
        <div className="col-sm-10">

          <div className="row">
            <div className="col-sm-4" align="center">
              <h1>238</h1>
              Buckets
            </div>          
            <div className="col-sm-4" align="center">
              <h1>375 TB</h1>
              Capacity
            </div>          
            <div className="col-sm-4">
            <Line data={{
  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'My First dataset',
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
              Usage VS Last Month
            </div>          
          </div>
          <div className="row">
            <div className="col-sm-4" align="center">
              <h1>1.5 TB</h1>
              Egress this Month
            </div>          
            <div className="col-sm-4">
              <Line data={{
  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'My First dataset',
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
              Network
            </div>          
            <div className="col-sm-4">
            <Line data={{
  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'My First dataset',
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
              Egress VS Last Month
            </div>          
          </div>
          <div className="row">
            <div className="col-sm-4" align="center">
              <h1>20 TB</h1>
              Ingress this month
            </div>          
            <div className="col-sm-4">
              <Bar
                data={{
                  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                  datasets: [
                    {
                      label: 'My First dataset',
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
              Year of Date
            </div>          
            <div className="col-sm-4">
            <Bar
                data={{
                  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
                  datasets: [
                    {
                      label: 'My First dataset',
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
              Month on Month
            </div>          
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
)(Dashboard);