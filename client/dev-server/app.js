const webpack = require('webpack');
const WebpackDevServer = require('webpack-dev-server');
const config = require('../webpack.config.client');

const PROXY_BACKEND = process.env.PROXY_BACKEND || 'http://127.0.0.1:5000';
const PORT = process.env.PORT || 3000;

config.entry['entry-client'].unshift(
  'react-hot-loader/patch',
  'webpack-dev-server/client',
  'webpack/hot/only-dev-server'
);

config.plugins.unshift(new webpack.HotModuleReplacementPlugin());
config.output.publicPath = '/';

new WebpackDevServer(webpack(config), {
  publicPath: '/',
  hot: true,
  proxy: {
    '/': {
      target: PROXY_BACKEND,
    },
  },
}).listen(PORT, function(err) {
  if (err) {
    // crash
    throw err;
  } else {
    console.log(
      `server started; listening on http://127.0.0.1:${this.address().port}`
    );
  }
});
