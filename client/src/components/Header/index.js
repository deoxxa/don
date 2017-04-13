// @flow

import classNames from 'classnames';
import React from 'react';
import { connect } from 'react-redux';
import { NavLink } from 'react-router-dom';

import FontAwesome from 'components/FontAwesome';
import HeaderLeft from 'components/HeaderLeft';
import HeaderRight from 'components/HeaderRight';
import Inner from 'components/Inner';

import styles from './styles.css';

const Header = ({ authenticationLoading, publicTimelineLoading }) => (
  <header className={styles.header}>
    <div
      className={classNames(styles.loading, {
        [styles.loadingActive]: authenticationLoading || publicTimelineLoading,
      })}
    />

    <Inner className={styles.inner}>
      <HeaderLeft />
      <HeaderRight />
    </Inner>
  </header>
);

export default connect(({
  authentication: { loading: authenticationLoading },
  publicTimeline: { loading: publicTimelineLoading },
}) => ({ authenticationLoading, publicTimelineLoading }))(Header);
