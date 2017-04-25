// @flow

import React, { Component } from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';
import { NavLink } from 'react-router-dom';

import sort, { reverseCompare } from 'lib/sort';

import type { State as AuthenticationState } from 'ducks/authentication';
import { publicTimelineAdd, publicTimelineFetch } from 'ducks/publicTimeline';
import type {
  ASActivity,
  State as PublicTimelineState,
} from 'ducks/publicTimeline';

import TimelineActivity from 'components/TimelineActivity';

import styles from './styles.css';

class Home extends Component {
  props: {
    authentication: AuthenticationState,
    publicTimeline: PublicTimelineState,
    publicTimelineFetch: () => Promise<void>,
    publicTimelineAdd: (activity: ASActivity) => void,
  };

  state: {
    selected: ?string,
  };

  constructor() {
    super();

    this.state = { selected: null };
  }

  componentDidMount() {
    const {
      publicTimeline: { activities },
      publicTimelineAdd,
      publicTimelineFetch,
    } = this.props;

    if (!Array.isArray(activities)) {
      publicTimelineFetch();
    }

    if (!this.events) {
      this.events = new EventSource('/api/feed');
      this.events.addEventListener('activity', ({ data }: { data: string }) => {
        try {
          publicTimelineAdd(JSON.parse(data));
        } catch (e) {
          /* nothing */
        }
      });
    }
  }

  componentWillUnmount() {
    if (this.events) {
      this.events.close();
      this.events = null;
    }
  }

  render() {
    const {
      publicTimeline: { activities },
      authentication: { user },
    } = this.props;

    const { selected } = this.state;

    return (
      <div className={styles.outer}>
        <div className={styles.activities}>
          {user
            ? <form
                className={styles.post}
                target="/post"
                onSubmit={ev => {
                  ev.preventDefault();
                  ev.target.reset();
                }}
              >
                <input
                  className={styles.input}
                  placeholder="It's good to post things."
                />
              </form>
            : null}

          <div>
            {sort(activities, 'time', reverseCompare).map(activity => (
              <TimelineActivity
                key={activity.id}
                activity={activity}
                selected={selected === activity.id}
                onClick={() => {
                  if (activity.id === selected) {
                    this.setState(() => ({ selected: null }));
                  } else {
                    this.setState(() => ({ selected: activity.id }));
                  }
                }}
              />
            ))}
          </div>
        </div>
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

export default connect(
  ({
    publicTimeline,
    authentication,
  }: {
    publicTimeline: PublicTimelineState,
    authentication: AuthenticationState,
  }) => ({ authentication, publicTimeline }),
  {
    publicTimelineAdd,
    publicTimelineFetch,
  }
)(Home);
