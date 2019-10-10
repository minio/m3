import { createSelector } from 'reselect'

const getSection = (state) => state.Dashboard.section;

export const getSelectedSection = createSelector(
  [ getSection ],
  (section) => {
    return section;
  },
);