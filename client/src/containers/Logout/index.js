// @flow

import React from 'react';
import { connect } from 'react-redux';

import { authenticationLogout } from 'ducks/authentication';
import type { State as AuthenticationState } from 'ducks/authentication';

import styles from './styles.css';

const Logout = (
  {
    authentication: { user },
    authenticationLogout,
    history,
  }: {
    authentication: AuthenticationState,
    authenticationLogout: () => Promise<void>,
    history: { push: (path: string) => void },
  }
) => (
  <div>
    {!user
      ? <h1>You are already logged out.</h1>
      : <form
          className={styles.form}
          action="/logout"
          method="post"
          onSubmit={ev => {
            ev.preventDefault();

            authenticationLogout().then(() => {
              history.push('/');
            });
          }}
        >
          <fieldset className={styles.fields}>
            <legend>Log Out</legend>

            <div className={styles.field}>
              <input type="submit" value="Log out" />
            </div>
          </fieldset>
        </form>}
  </div>
);

export default connect(({ authentication }) => ({ authentication }), {
  authenticationLogout,
})(Logout);
