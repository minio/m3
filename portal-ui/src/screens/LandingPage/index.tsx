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
import CssBaseline from "@material-ui/core/CssBaseline";
import AppBar from "@material-ui/core/AppBar/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import Link from "@material-ui/core/Link";
import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Box from "@material-ui/core/Box";
import {makeStyles} from '@material-ui/core/styles';
import { useHistory } from 'react-router'
import Copyright from "../../components/Copyright";


const useStyles = makeStyles(theme => ({
    '@global': {
        body: {
            backgroundColor: theme.palette.common.white,
        },
        ul: {
            margin: 0,
            padding: 0,
        },
        li: {
            listStyle: 'none',
        },
    },
    appBar: {
        borderBottom: `1px solid ${theme.palette.divider}`,
    },
    toolbar: {
        flexWrap: 'wrap',
    },
    toolbarTitle: {
        flexGrow: 1,
    },
    link: {
        margin: theme.spacing(1, 1.5),
    },
    heroContent: {
        padding: theme.spacing(8, 0, 6),
    },
    cardHeader: {
        backgroundColor: theme.palette.grey[200],
    },
    cardPricing: {
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'baseline',
        marginBottom: theme.spacing(2),
    },
    footer: {
        borderTop: `1px solid ${theme.palette.divider}`,
        marginTop: theme.spacing(8),
        paddingTop: theme.spacing(3),
        paddingBottom: theme.spacing(3),
        [theme.breakpoints.up('sm')]: {
            paddingTop: theme.spacing(6),
            paddingBottom: theme.spacing(6),
        },
    },
    modal: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
    },
    paper: {
        backgroundColor: theme.palette.background.paper,
        border: '2px solid #000',
        boxShadow: theme.shadows[5],
        padding: theme.spacing(2, 4, 3),
    },
}));


const footers = [
    {
        title: 'Company',
        description: ['Team', 'History', 'Contact us', 'Locations'],
    },
    {
        title: 'Features',
        description: ['Cool stuff', 'Random feature', 'Team feature', 'Developer stuff', 'Another one'],
    },
    {
        title: 'Resources',
        description: ['Resource', 'Resource name', 'Another resource', 'Final resource'],
    },
    {
        title: 'Legal',
        description: ['Privacy policy', 'Terms of use'],
    },
];


const Landing: React.FC = () => {
    const classes = useStyles();
    const { push } = useHistory()

    return (
        <React.Fragment>
            <CssBaseline/>
            <AppBar position="static" color="default" elevation={0} className={classes.appBar}>
                <Toolbar className={classes.toolbar}>
                    <Typography variant="h6" color="inherit" noWrap className={classes.toolbarTitle}>
                        Acme Storage
                    </Typography>
                    <nav>
                        <Link variant="button" color="textPrimary" href="#" className={classes.link}>
                            Features
                        </Link>
                        <Link variant="button" color="textPrimary" href="#" className={classes.link}>
                            Enterprise
                        </Link>
                        <Link variant="button" color="textPrimary" href="#" className={classes.link}>
                            Support
                        </Link>
                    </nav>
                    <Button href="#" color="primary" variant="outlined" className={classes.link} onClick={(e)=>push('/login')}>
                        Login
                    </Button>
                </Toolbar>
            </AppBar>
            {/* Hero unit */}
            <Container maxWidth="sm" component="main" className={classes.heroContent}>
                <Typography component="h1" variant="h2" align="center" color="textPrimary" gutterBottom>
                    World-Class Storage
                </Typography>
                <Typography variant="h5" align="center" color="textSecondary" component="p">
                    Quickly setup world-class, s3-compatible, cost-efficient storage that uses MinIO underneath.
                </Typography>
            </Container>
            {/* End hero unit */}
            <Container maxWidth="md" component="main">
                <Typography variant="body1" align="center" color="textPrimary" component="p">
                    Setting up the access takes a few minutes, you just need a credit card.
                </Typography>
                <Typography variant="body1" align="center" color="textPrimary" component="p">
                    <Button variant="contained" size="large" color="primary" onClick={(e)=>push('/signup')}>
                        Sign Up!
                    </Button>

                </Typography>
            </Container>
            {/* Footer */}
            <Container maxWidth="md" component="footer" className={classes.footer}>
                <Grid container spacing={4} justify="space-evenly">
                    {footers.map(footer => (
                        <Grid item xs={6} sm={3} key={footer.title}>
                            <Typography variant="h6" color="textPrimary" gutterBottom>
                                {footer.title}
                            </Typography>
                            <ul>
                                {footer.description.map(item => (
                                    <li key={item}>
                                        <Link href="#" variant="subtitle1" color="textSecondary">
                                            {item}
                                        </Link>
                                    </li>
                                ))}
                            </ul>
                        </Grid>
                    ))}
                </Grid>
                <Box mt={7}>
                    <Copyright/>
                </Box>
            </Container>
            {/* End footer */}
        </React.Fragment>
    );
}

export default Landing;
