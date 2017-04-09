import React from 'react';
import { renderToString } from 'react-dom/server';
import { StaticRouter } from 'react-router';

import Root from './root';
import { setupAxios, setupStore } from './setup';

module.exports = function main(location, initialState) {
  const store = setupStore(initialState);

  return renderToString(
    <StaticRouter location={location}>
      <Root store={store} />
    </StaticRouter>
  );
};
