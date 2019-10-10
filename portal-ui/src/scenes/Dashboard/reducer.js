import { SELECT } from './actions';

const initialState = {
  section: 'metrics',
};

export const reducer = (state = initialState, action) => {
	switch (action.type) {
    case SELECT:
    	return {
        ...state,
        section: action.payload,
      }
		default:
			return state;
	}
}