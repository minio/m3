import React from 'react';
import { compose } from 'recompose'
import { connect } from 'react-redux';

import './styles.css';

function LandingPage(props) {
  return (
    <div className="LandingPage">
      <div className="container">
        <div className="row">
          <div className="col-sm-12">
            <p className="text-center">
              <h1>ACME Cloud Storage</h1>
            </p>
          </div>
        </div>
        <div className="row">
          <div className="col-sm-7">
            <p className="text-justify">
            Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi sagittis vulputate nulla, in egestas orci tincidunt et. Duis interdum, mi in faucibus dapibus, lectus massa accumsan ex, vel placerat purus ante ac ex. Nulla facilisi. Phasellus at purus sed tortor maximus vehicula. Suspendisse eu luctus enim, non consectetur magna. Fusce at tempor libero, vel maximus leo. In pretium eu metus id rhoncus. Nulla feugiat, leo a dictum fermentum, ipsum massa vulputate odio, et euismod nibh ligula at ipsum. Morbi finibus sollicitudin nisi eu iaculis. In efficitur dui imperdiet arcu molestie tincidunt. Suspendisse sit amet dolor velit. Duis vitae dui id ipsum elementum mollis vel at justo. Nam libero neque, congue dictum libero sed, imperdiet eleifend sem. Nullam molestie risus vel nunc faucibus congue. Suspendisse non nunc interdum, pellentesque felis ut, ornare lectus.
Vestibulum porttitor rhoncus urna, eget cursus nibh sagittis ut. Aliquam eget suscipit lacus, vitae facilisis ligula. Cras eget massa elementum augue tempus faucibus. Praesent dictum augue a nisi interdum, at rhoncus purus vestibulum. Donec faucibus mi quis sollicitudin rhoncus. Mauris venenatis sapien sit amet dui vestibulum eleifend. Donec ut bibendum ex. Donec condimentum eleifend nisl, quis scelerisque augue varius sed. Aenean gravida tincidunt mi, nec sodales erat faucibus ac.
            </p>
          </div>
          <div className="col-sm-5">
            <img alt="" className="img-responsive col-sm-12" src="https://www.jamesmyersco.com/wp-content/uploads/2015/10/international-finance-corporation-building.jpg" />
          </div>
        </div>
        <div className="row">
          <div className="col-sm-12">
            <p className="text-justify">
              Maecenas nec est quis ante interdum malesuada. Sed est mauris, mattis nec nulla in, vestibulum efficitur orci. Donec facilisis quam et urna commodo, ac fringilla purus feugiat. Phasellus non ante orci. Nunc quam lacus, suscipit laoreet varius nec, lacinia eget lorem. Sed lectus elit, luctus quis vehicula quis, aliquet a sem. Proin volutpat, arcu in porttitor eleifend, enim risus elementum est, ut vulputate quam nulla in ligula. Quisque at dui semper, porttitor nunc ac, dapibus tellus. Nulla imperdiet, tortor ac molestie posuere, nibh nulla placerat dolor, nec pretium est sem ornare lorem. Mauris laoreet placerat nunc, et consequat ante commodo non. Donec cursus sagittis efficitur. Nulla consectetur malesuada mollis. Vestibulum et leo dolor. Aenean euismod volutpat quam, aliquam ultricies nisi sagittis vel. Etiam interdum tempor pharetra. Etiam hendrerit lorem lacus, vel sollicitudin quam convallis id. Nam ultricies facilisis nulla sed pretium. Suspendisse tincidunt quis neque in pharetra. Vivamus non rutrum diam, eu posuere lacus. Quisque dignissim magna non nulla consectetur, id egestas tortor placerat. Integer enim sapien, pretium eget dictum vitae, molestie eget sapien.
            </p>
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
)(LandingPage);