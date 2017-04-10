import React, { Component } from 'react';
import { connect } from 'react-redux';
import { Link } from 'react-router';

import FontAwesome from 'components/FontAwesome';
import PublicTimelinePost from 'components/PublicTimelinePost';

import styles from './styles.css';

export class Home extends Component {
  render() {
    const { publicTimeline: { posts = [] } } = this.props;

    return (
      <div>
        <form className={styles.form} method="get" action="/find-feed">
          <label htmlFor={styles.userInput}>
            Find a user anywhere in the fediverse!
          </label>

          <div>
            <input
              id={styles.userInput}
              name="user"
              type="text"
              placeholder="e.g. your-username@your-provider.com"
              required
            />

            <input type="submit" value="👀" />
          </div>
        </form>

        <p class={styles.blurb}>
          This is a <em>ridiculously</em> simple, read-only StatusNet node.
          Mostly an experiment.
          <a href="https://www.fknsrs.biz/p/don">Source code is available</a>.
        </p>

        <h1>
          Here are some posts from the public timeline!
        </h1>

        {(posts || [])
          .map(post => <PublicTimelinePost key={post.id} post={post} />)}
      </div>
    );
  }
}

export default connect(({ publicTimeline }) => ({ publicTimeline }))(Home);
