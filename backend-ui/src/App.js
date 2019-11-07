// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
import React from 'react';
import './App.css';

function App() {
    const {counter} = this.props
    return (
        <div className="App">
            <header className="App-header">
                <h1>m3</h1>
                <p>
                    {counter}
                </p>
                <p>
                    Coming soon
                </p>
            </header>
        </div>
    );
}

const mapStateToProps = state => {
    return {
        counter: state.counter,
    }
}

const mapDispatchToProps = dispatch => {
    return {
        showAddNodeModal: () => dispatch(showAddNodeModal())
    }
}


const AppController = connect(
    mapStateToProps,
    mapDispatchToProps
)(DeploymentsHeader)

export { AppController as App }


// export default App;
