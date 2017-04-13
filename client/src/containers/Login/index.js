// @flow

import React from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';
import { Link, NavLink } from 'react-router-dom';
import URLSearchParams from 'url-search-params';

import { authenticationLogin } from 'ducks/authentication';
import type { State as AuthenticationState } from 'ducks/authentication';

import FontAwesome from 'components/FontAwesome';
import PublicTimelinePost from 'components/PublicTimelinePost';

import styles from './styles.css';

const Login = (
  {
    authenticationLogin,
    authentication: { user },
    location,
    history,
  }: {
    authenticationLogin: (username: string, password: string) => Promise<void>,
    authentication: AuthenticationState,
    location: { search: string },
    history: { push: (path: string) => void },
  }
) => (
  <div>
    {user
      ? <h1 className={styles.heading}>
          You are logged in as {user.username}.
        </h1>
      : <form
          className={styles.form}
          action="/login"
          method="post"
          onSubmit={ev => {
            ev.preventDefault();

            authenticationLogin(
              ev.target[1].value,
              ev.target[2].value
            ).then(() => {
              history.push(
                new URLSearchParams(location.search).get('return_to') || '/'
              );
            });
          }}
        >
          <fieldset className={styles.fields}>
            <legend>Log In</legend>

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
              <input type="submit" value="Log in" />
            </div>

            <hr className={styles.line} />

            <div className={styles.link}>
              <span>
                No account? <Link to="/register">Register here.</Link>
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
  authenticationLogin,
})(Login);
