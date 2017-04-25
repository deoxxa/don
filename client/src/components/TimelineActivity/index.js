// @flow

import classNames from 'classnames';
import React, { Component } from 'react';
import TimeAgo from 'react-timeago';

import type { Activity } from 'ducks/publicTimeline';

import styles from './styles.css';

// const verbs = {
//   'http://activitystrea.ms/schema/1.0/follow': 'followed',
//   'http://activitystrea.ms/schema/1.0/post': 'posted',
//   'http://activitystrea.ms/schema/1.0/share': 'shared',
//   'http://activitystrea.ms/schema/1.0/favorite': 'favorited',
// };

// const Verb = ({ className, verb }: { className?: ?string, verb: string }) => (
//   <span className={className}>{verbs[verb] || verb}</span>
// );

export default class TimelineActivity extends Component {
  props: { activity: Activity, selected: boolean, onClick: () => void };
  state: { mounted: boolean };

  constructor() {
    super();

    this.state = { mounted: false };
  }

  componentDidMount() {
    this.setState(() => ({ mounted: true }));
  }

  render() {
    const { activity, selected, onClick } = this.props;
    const { mounted } = this.state;

    return (
      <div
        className={classNames(styles.outer, {
          [styles.selected]: selected,
          [styles.deselected]: !selected,
        })}
        onClick={(ev: MouseEvent) => {
          for (let el = ev.target; el.parentNode; el = el.parentNode) {
            switch (el.tagName) {
              case 'A':
                return;
            }
          }

          onClick();
        }}
      >
        {activity.actor
          ? <a
              href={`/find-feed?user=${activity.actor.id}`}
              className={styles.avatar}
            >
              <img
                height="60"
                className={styles.avatarImage}
                src={activity.actor.avatar}
              />
            </a>
          : null}

        <div className={styles.body}>
          <div className={styles.header}>
            {activity.actor
              ? <a
                  className={styles.displayName}
                  href={`/find-feed?user=${activity.actor.id}`}
                  title={activity.actor.id}
                >
                  {activity.actor.displayName}
                </a>
              : <span className={styles.displayName}>
                  No Username Available
                </span>}

            <a className={styles.time} href={activity.permalink}>
              {mounted
                ? <TimeAgo
                    date={activity.time}
                    formatter={(value: number, unit: string, suffix: string) =>
                      `${value} ${unit}${value === 1 ? '' : 's'} ${suffix}`}
                  />
                : <span>at <time>{activity.time}</time></span>}
            </a>
          </div>

          {activity.object.content
            ? <div
                className={styles.content}
                dangerouslySetInnerHTML={{ __html: activity.object.content }}
              />
            : null}

          <div className={styles.footer}>
            {activity.inReplyToURL
              ? <span className={styles.inReplyTo}>
                  in reply to <a href={activity.inReplyToURL}>this</a>
                </span>
              : null}

            {activity.object.permalink !== activity.permalink &&
              activity.object.permalink !== activity.inReplyToURL
              ? <span className={styles.from}>
                  from <a href={activity.object.permalink}>here</a>
                </span>
              : null}
          </div>
        </div>

      </div>
    );
  }
}
