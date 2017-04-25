// @flow

import React, { Component } from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';
import { NavLink } from 'react-router-dom';
import URLSearchParams from 'url-search-params';

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
    location: { search: ?string },
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
      location,
      publicTimeline: { activities },
      publicTimelineFetch,
    } = this.props;

    if (!Array.isArray(activities)) {
      const params = new URLSearchParams(location.search || '');

      publicTimelineFetch({
        before: params.get('before'),
        after: params.get('after'),
        q: params.get('q'),
      });
    }

    this.connectEvents(this.props);
  }

  componentWillReceiveProps(nextProps) {
    const { publicTimelineFetch } = nextProps;

    const oldParams = new URLSearchParams(this.props.location.search || '');
    const newParams = new URLSearchParams(nextProps.location.search || '');

    if (oldParams.get('q') !== newParams.get('q')) {
      publicTimelineFetch({ q: newParams.get('q') });
      this.reconnectEvents(nextProps);
    }
  }

  componentWillUnmount() {
    this.disconnectEvents();
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

  _feed: ?Object;

  reconnectEvents(props) {
    this.disconnectEvents();
    this.connectEvents(props);
  }

  connectEvents(props) {
    const { location: { search = '' }, publicTimelineAdd } = props;

    const params = new URLSearchParams(search);

    this._feed = new EventSource(`/api/feed?q=${params.get('q') || ''}`);

    this._feed.addEventListener('activity', ({ data }: { data: string }) => {
      try {
        publicTimelineAdd(JSON.parse(data));
      } catch (e) {
        /* nothing */
      }
    });
  }

  disconnectEvents() {
    if (this._feed) {
      this._feed.close();
      this._feed = null;
    }
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
