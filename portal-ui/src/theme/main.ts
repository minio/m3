import { createMuiTheme } from '@material-ui/core';

const theme = createMuiTheme({
  palette: {
    primary: {
      light: '#757ce8',
      main: '#3f50b5',
      dark: '#002884',
      contrastText: '#fff',
    },
    secondary: {
      light: '#ff7961',
      main: '#f44336',
      dark: '#ba000d',
      contrastText: '#000',
    },
    error: {
      light: '#e03a48',
      main: '#dc1f2e',
      contrastText: '#ffffff',
    },
    grey: {
      100: '#f0f0f0',
      200: '#e6e6e6',
      300: '#cccccc',
      400: '#999999',
      500: '#8c8c8c',
      600: '#737373',
      700: '#666666',
      800: '#4d4d4d',
      900: '#333333',
    },
    background: {
      default: '#fafafa',
    },
  },
});

export default theme;
