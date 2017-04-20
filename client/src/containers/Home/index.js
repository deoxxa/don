// @flow

import React, { Component } from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';
import { NavLink } from 'react-router-dom';

import { publicTimelineFetch } from 'ducks/publicTimeline';
import type { State as PublicTimelineState } from 'ducks/publicTimeline';

import PublicTimelineActivity from 'components/PublicTimelineActivity';

import styles from './styles.css';

class Home extends Component {
  props: {
    publicTimeline: PublicTimelineState,
    publicTimelineFetch: () => Promise<void>,
  };

  componentDidMount() {
    const { publicTimeline: { activities }, publicTimelineFetch } = this.props;

    if (!Array.isArray(activities)) {
      publicTimelineFetch();
    }
  }

  render() {
    const { publicTimeline: { activities } } = this.props;

    return (
      <div>
        <p className={styles.blurb}>
          This is a <em>ridiculously</em> simple, read-only StatusNet node.
          Mostly an experiment.
          {' '}
          <a href="https://www.fknsrs.biz/p/don">Source code is available</a>.
        </p>

        <h1>
          Here are some posts from the public timeline!
        </h1>

        {(activities || [])
          .map(activity => <PublicTimelineActivity key={activity.id} activity={activity} />)}
      </div>
    );
  }
}

Home.Link = withRouter(
  connect(
    ({ publicTimeline }: { publicTimeline: PublicTimelineState }) => ({
      publicTimeline,
    }),
    { publicTimelineFetch }
  )(({
    history: { push },
    publicTimeline: { activities },
    publicTimelineFetch,
    children,
    ...rest
  }: {
    publicTimeline: PublicTimelineState,
    history: { push: (path: string) => void },
    publicTimelineFetch: () => Promise<void>,
    children?: React.Children,
  }) => (
    <NavLink
      {...rest}
      onClick={ev => {
        ev.preventDefault();

        if (Array.isArray(activities)) {
          push('/');
        } else {
          publicTimelineFetch().then(() => push('/'));
        }
      }}
    >
      {children}
    </NavLink>
  ))
);

export default connect(({ publicTimeline }) => ({ publicTimeline }), {
  publicTimelineFetch,
})(Home);
