// @flow

import React, { Component } from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';
import { Link, NavLink } from 'react-router-dom';
import URLSearchParams from 'url-search-params';

import { authenticationRegister } from 'ducks/authentication';
import type { State as AuthenticationState } from 'ducks/authentication';

import FontAwesome from 'components/FontAwesome';
import PublicTimelinePost from 'components/PublicTimelinePost';

import styles from './styles.css';

const Register = (
  {
    authentication: { user },
    location,
  }: {
    authentication: AuthenticationState,
    location: { search: string },
  }
) => (
  <div>
    {user
      ? <h1>You are logged in as {user.username}.</h1>
      : <form className={styles.form} action="/register" method="post">
          <fieldset className={styles.fields}>
            <legend>Register</legend>

            <div className={styles.field}>
              <label className={styles.label} htmlFor={styles.usernameInput}>
                Username:
              </label>
              <input
                id={styles.usernameInput}
                name="username"
                type="text"
                required
              />
            </div>

            <div className={styles.field}>
              <label className={styles.label} htmlFor={styles.emailInput}>
                Email:
              </label>
              <input id={styles.emailInput} name="email" type="text" required />
            </div>

            <div className={styles.field}>
              <label className={styles.label} htmlFor={styles.passwordInput}>
                Password:
              </label>
              <input
                id={styles.passwordInput}
                name="password"
                type="password"
                required
              />
            </div>

            <div className={styles.field}>
              <input type="submit" value="Register" />
            </div>

            <hr className={styles.line} />

            <div className={styles.field}>
              <span>
                Already registered? <Link to="/login">Log in here.</Link>
              </span>
            </div>

            <input
              type="hidden"
              name="return_to"
              value={
                new URLSearchParams(location.search).get('return_to') || '/'
              }
            />
          </fieldset>
        </form>}
  </div>
);

export default connect(({ authentication }) => ({ authentication }), {
  authenticationRegister,
})(Register);
