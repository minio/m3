import React from 'react';
import NavigationBar from './components/NavigationBar';
import LandingPage from './scenes/LandingPage';
import './App.css';

function App() {
  return (
    <div className="App">
      <NavigationBar />
      <LandingPage />
    </div>
  );
}

export default App;
