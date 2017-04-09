const ExtractTextPlugin = require('extract-text-webpack-plugin');
const UnusedFilesWebpackPlugin = require('unused-files-webpack-plugin').UnusedFilesWebpackPlugin;
const webpack = require('webpack');

module.exports = require('./webpack.config.common');

module.exports.entry = {
  'entry-client': ['./src/entry-client'],
  vendor: [
    'axios',
    'classnames',
    'classnames/bind',
    'font-awesome/css/font-awesome.css',
    'nonmutable',
    'qs',
    'react',
    'react-dom',
    'react-redux',
    'react-router-dom',
    'redbox-react',
    'redux',
    'redux-actions',
    'redux-localstorage',
    'redux-logger',
    'redux-thunk',
    'uuid',
  ],
};

if (process.env.NODE_ENV === 'production') {
  module.exports.plugins.push(
    new webpack.optimize.CommonsChunkPlugin({
      name: 'vendor',
      minChunks: Infinity,
    })
  );

  module.exports.plugins.push(
    new UnusedFilesWebpackPlugin({
      pattern: 'src/**/*.*',
      globOptions: { ignore: ['src/**/__tests__/*.js', 'src/entry-*.js'] },
      failOnUnused: true,
    })
  );
}

module.exports.plugins.push(new ExtractTextPlugin('[name]-styles.css'));

module.exports.module.rules.forEach(rule => {
  if (rule.use.indexOf('style-loader') === 0) {
    rule.use = ExtractTextPlugin.extract({
      fallback: rule.use[0],
      use: rule.use.slice(1),
    });
  }
});
