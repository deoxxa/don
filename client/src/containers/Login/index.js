// @flow

import React from 'react';
import { connect } from 'react-redux';
import { Link } from 'react-router-dom';
import URLSearchParams from 'url-search-params';

import { authenticationLogin } from 'ducks/authentication';
import type { State as AuthenticationState } from 'ducks/authentication';

import styles from './styles.css';

const Login = ({
  authenticationLogin,
  authentication: { error, user },
  location,
  history,
}: {
  authenticationLogin: (username: string, password: string) => Promise<void>,
  authentication: AuthenticationState,
  location: { search: string },
  history: { push: (path: string) => void },
}) => {
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
            action={`/login?return_to=${returnTo}`}
            onSubmit={ev => {
              ev.preventDefault();

              authenticationLogin(ev.target[1].value, ev.target[2].value).then(
                () => history.push(returnTo),
                () => {}
              );
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
                  No account?
                  {' '}
                  <Link to={`/register?return_to=${returnTo}`}>
                    Register here.
                  </Link>
                </span>
              </div>
            </fieldset>
          </form>}
    </div>
  );
};

export default connect(({ authentication }) => ({ authentication }), {
  authenticationLogin,
})(Login);
