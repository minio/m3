import { createStore, applyMiddleware, combineReducers } from 'redux';
import thunk from 'redux-thunk';

const appReducer = combineReducers({

});

export default function configureStore() {
 return createStore(
    appReducer,
    applyMiddleware(thunk),
 );
};
