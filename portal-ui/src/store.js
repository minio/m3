import { createStore, applyMiddleware, combineReducers } from 'redux';
import thunk from 'redux-thunk';
import { reducer as DashboardReducer } from './scenes/Dashboard/reducer';

const appReducer = combineReducers({
  Dashboard: DashboardReducer,
});

export default function configureStore() {
 return createStore(
    appReducer,
    applyMiddleware(thunk),
 );
};
