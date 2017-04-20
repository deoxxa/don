// @flow

import React, { Component } from 'react';
import TimeAgo from 'react-timeago';

import type { Activity } from 'ducks/publicTimeline';

import styles from './styles.css';

const verbs = {
  'http://activitystrea.ms/schema/1.0/follow': 'followed',
  'http://activitystrea.ms/schema/1.0/post': 'posted',
  'http://activitystrea.ms/schema/1.0/share': 'shared',
};

const Verb = ({ verb }: { verb: string }) => <span>{verbs[verb] || verb}</span>;

export default class PublicTimelineActivity extends Component {
  props: { activity: Activity };
  state: { mounted: boolean };

  constructor() {
    super();

    this.state = { mounted: false };
  }

  componentDidMount() {
    this.setState(() => ({ mounted: true }));
  }

  render() {
    const { activity } = this.props;
    const { mounted } = this.state;

    return (
      <div>
        <h3>
          {activity.actor
            ? <span>
                <img src={activity.actor.avatar} width="32" />

                <a href={activity.actor.permalink} title={activity.actor.id}>
                  {activity.actor.displayName}
                </a>
              </span>
            : <span>No Username Available</span>}

          {' '}

          (<span>
            <Verb verb={activity.verb} />

            {' '}

            <a href={activity.object.permalink}>
              {mounted
                ? <TimeAgo
                    date={activity.time}
                    formatter={(value: number, unit: string, suffix: string) =>
                      `${value} ${unit}${value === 1 ? '' : 's'} ${suffix}`}
                  />
                : <time>{activity.time}</time>}
            </a>

            {activity.inReplyToURL
              ? <span>
                  {' '} in reply to <a href={activity.inReplyToURL}>this</a>
                </span>
              : null}

            {activity.object.permalink !== activity.permalink
              ? <span>
                  {' '} from <a href={activity.object.permalink}>here</a>
                </span>
              : null}
          </span>)
        </h3>

        {activity.object.content
          ? <div
              className={styles.content}
              dangerouslySetInnerHTML={{ __html: activity.object.content }}
            />
          : null}
      </div>
    );
  }
}
