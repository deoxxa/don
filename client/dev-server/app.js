const webpack = require('webpack');
const WebpackDevServer = require('webpack-dev-server');
const config = require('../webpack.config');

const PROXY_API = process.env.PROXY_API || 'http://127.0.0.1:3000';
const PORT = process.env.PORT || 3000;

config.entry.browser.unshift(
  'react-hot-loader/patch',
  'webpack-dev-server/client',
  'webpack/hot/only-dev-server'
);

config.plugins.unshift(new webpack.HotModuleReplacementPlugin());

config.module.loaders.forEach(loader => {
  const a = loader.loaders;

  if (a.includes('css?modules')) {
    a[1] += '&localIdentName=[path][name]---[local]---[hash:base64:5]';
  }
});

new WebpackDevServer(webpack(config), {
  contentBase: __dirname,
  publicPath: config.output.publicPath,
  hot: true,
  historyApiFallback: true,
  proxy: {
    '/api/*': {
      target: PROXY_API,
    },
  },
}).listen(PORT, err => {
  if (err) {
    // crash
    throw err;
  } else {
    console.log('server started');
  }
});
