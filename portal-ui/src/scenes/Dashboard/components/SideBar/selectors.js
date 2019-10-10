import { createSelector } from 'reselect'

const getSection = (state) => state.Dashboard.section;

export const getSelecteSection = createSelector(
  [ getSection ],
  (section) => {
    return section;
  },
);