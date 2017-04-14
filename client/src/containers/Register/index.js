// @flow

import React from 'react';
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
    authenticationRegister,
    authentication: { error, user },
    location,
    history,
  }: {
    authenticationRegister: (
      email: string,
      username: string,
      password: string
    ) => Promise<void>,
    authentication: AuthenticationState,
    location: { search: string },
    history: { push: (path: string) => void },
  }
) => {
  const returnTo = new URLSearchParams(location.search).get('return_to') || '/';

  return (
    <div>
      {error ? <h3 className={styles.error}>{error}</h3> : null}

      {user
        ? <h1 className={styles.heading}>
            You are logged in as {user.username}.
          </h1>
        : <form
            className={styles.form}
            method="post"
            action={`/register?return_to=${returnTo}`}
            onSubmit={ev => {
              ev.preventDefault();

              authenticationRegister(
                ev.target[1].value,
                ev.target[2].value,
                ev.target[3].value
              ).then(() => history.push(returnTo), err => null);
            }}
          >
            <fieldset className={styles.fields}>
              <legend>Register</legend>

              <div className={styles.field}>
                <label className={styles.label} htmlFor={styles.emailInput}>
                  Email:
                </label>
                <input
                  id={styles.emailInput}
                  name="email"
                  type="email"
                  required
                />
              </div>

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
                <input type="submit" value="Register" />
              </div>

              <hr className={styles.line} />

              <div className={styles.field}>
                <span>
                  Already registered?
                  {' '}
                  <Link to={`/login?return_to=${returnTo}`}>Log in here.</Link>
                </span>
              </div>
            </fieldset>
          </form>}
    </div>
  );
};

export default connect(({ authentication }) => ({ authentication }), {
  authenticationRegister,
})(Register);
