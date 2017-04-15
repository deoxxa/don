// @flow

import React from 'react';
import { renderToString } from 'react-dom/server';
import { StaticRouter } from 'react-router';

import type { State } from 'ducks';

import Root from './root';
import { setupStore } from './setup';

module.exports = function main(location: string, initialStateJSON: string) {
  let initialState: ?State = null;
  if (initialStateJSON) {
    try {
      initialState = JSON.parse(initialStateJSON);
    } catch (e) {
      // do nothing
    }
  }

  const store = setupStore(initialState);

  return renderToString(
    <StaticRouter location={location}>
      <Root store={store} />
    </StaticRouter>
  );
};
