// @flow

import React from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';
import { NavLink } from 'react-router-dom';
import URLSearchParams from 'url-search-params';

import { serialiseForm } from 'lib/serialiseForm';

import type { User } from 'ducks/authentication';

import FontAwesome from 'components/FontAwesome';

import defaultAvatarURL from './default-avatar.png';
import styles from './styles.css';

const HeaderRight = (
  {
    user,
    history,
    location,
  }: {
    user: ?User,
    history: { push: (path: string) => void },
    location: { search: string },
  }
) => (
  <nav className={styles.nav}>
    <form
      action="/"
      method="get"
      className={styles.form}
      onSubmit={ev => {
        ev.preventDefault();

        history.push(serialiseForm(ev.target));
      }}
    >
      <input
        name="q"
        className={styles.search}
        type="text"
        placeholder="Search"
        defaultValue={new URLSearchParams(location.search).get('q')}
      />
    </form>

    {user
      ? <NavLink to="/logout" title={user.displayName}>
          <img
            className={styles.avatar}
            src={user.avatar || defaultAvatarURL}
          />
        </NavLink>
      : <NavLink
          className={styles.login}
          to="/login"
          title="Log in or register"
        >
          <FontAwesome icon="sign-in" />
        </NavLink>}
  </nav>
);

export default withRouter(
  connect(({ authentication: { user } }) => ({ user }))(HeaderRight)
);
