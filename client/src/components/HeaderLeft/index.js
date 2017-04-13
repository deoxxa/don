// @flow

import React from 'react';
import { NavLink } from 'react-router-dom';

import Home from 'containers/Home';

import FontAwesome from 'components/FontAwesome';

import styles from './styles.css';

const HeaderLeft = () => (
  <nav className={styles.nav}>
    <Home.Link
      className={styles.link}
      activeClassName={styles.active}
      exact
      to="/"
    >
      <FontAwesome className={styles.icon} icon="home" /> Home
    </Home.Link>
  </nav>
);

export default HeaderLeft;
