// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2020 MinIO, Inc.
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

import React from "react";
import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import CssBaseline from "@material-ui/core/CssBaseline";
import TextField from "@material-ui/core/TextField";
import Link from "@material-ui/core/Link";
import Grid from "@material-ui/core/Grid";
import Box from "@material-ui/core/Box";
import LockOutlinedIcon from "@material-ui/icons/LockOutlined";
import Typography from "@material-ui/core/Typography";
import { createStyles, Theme, withStyles } from "@material-ui/core/styles";
import Container from "@material-ui/core/Container";
import Copyright from "../../components/Copyright";
import request from "superagent";
import storage from "local-storage-fallback";
import history from "./../../history";
import { connect, ConnectedProps } from "react-redux";
import { userLoggedIn } from "../../actions";
import { SystemState } from "../../types";

const styles = (theme: Theme) =>
  createStyles({
    "@global": {
      body: {
        backgroundColor: theme.palette.common.white
      }
    },
    paper: {
      marginTop: theme.spacing(8),
      display: "flex",
      flexDirection: "column",
      alignItems: "center"
    },
    avatar: {
      margin: theme.spacing(1),
      backgroundColor: theme.palette.secondary.main
    },
    form: {
      width: "100%", // Fix IE 11 issue.
      marginTop: theme.spacing(3)
    },
    submit: {
      margin: theme.spacing(3, 0, 2)
    },
    errorBlock: {
      color: "red"
    }
  });

const mapState = (state: SystemState) => ({
  loggedIn: state.loggedIn
});

const connector = connect(mapState, { userLoggedIn });

// The inferred type will look like:
// {isOn: boolean, toggleOn: () => void}
type PropsFromRedux = ConnectedProps<typeof connector>;
type Props = PropsFromRedux & {};

interface LoginProps {
  userLoggedIn: typeof userLoggedIn;
  classes: any;
}

class Login extends React.Component<LoginProps> {
  state = {
    email: "",
    password: "",
    company: "",
    error: ""
  };

  formSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const url = "/api/v1/users/login";
    const { email, password, company } = this.state;
    request
      .post(url)
      .send({ email: email, password: password, company: company })
      .then((res: any) => {
        if (res.body.jwt_token) {
          // store the jwt token
          storage.setItem("token", res.body.jwt_token);
          return res.body.jwt_token;
        } else if (res.body.error) {
          // throw will be moved to catch block once bad login returns 403
          throw res.body.error;
        }
      })
      .then(() => {
        // push('/dashboard');
        this.props.userLoggedIn(true);
        history.push("/dashboard");
      })
      .catch(err => {
        this.setState({ error: err });
      });
  };

  render() {
    const { error, email, password, company } = this.state;
    const { classes } = this.props;
    return (
      <Container component="main" maxWidth="xs">
        <CssBaseline />
        <div className={classes.paper}>
          <Avatar className={classes.avatar}>
            <LockOutlinedIcon />
          </Avatar>
          <Typography component="h1" variant="h5">
            Acme Storage Login
          </Typography>
          <form className={classes.form} noValidate onSubmit={this.formSubmit}>
            <Grid container spacing={2}>
              {error !== "" && (
                <Grid item xs={12}>
                  <Typography
                    component="p"
                    variant="body1"
                    className={classes.errorBlock}
                  >
                    {`${error}`}
                  </Typography>
                </Grid>
              )}
              <Grid item xs={12}>
                <TextField
                  autoComplete="company_name"
                  name="company_name"
                  variant="outlined"
                  required
                  fullWidth
                  value={company}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    this.setState({ company: e.target.value })
                  }
                  id="company_name"
                  label="Company"
                  autoFocus
                />
              </Grid>

              <Grid item xs={12}>
                <TextField
                  variant="outlined"
                  required
                  fullWidth
                  id="email"
                  value={email}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    this.setState({ email: e.target.value })
                  }
                  label="Email Address"
                  name="email"
                  autoComplete="email"
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  variant="outlined"
                  required
                  fullWidth
                  value={password}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    this.setState({ password: e.target.value })
                  }
                  name="password"
                  label="Password"
                  type="password"
                  id="password"
                  autoComplete="current-password"
                />
              </Grid>
            </Grid>
            <Button
              type="submit"
              fullWidth
              variant="contained"
              color="primary"
              className={classes.submit}
            >
              Login
            </Button>
            <Grid container justify="flex-end">
              <Grid item>
                <Link href="#" variant="body2">
                  Forgot Password?
                </Link>
              </Grid>
            </Grid>
          </form>
        </div>
        <Box mt={5}>
          <Copyright />
        </Box>
      </Container>
    );
  }
}

export default connector(withStyles(styles)(Login));
