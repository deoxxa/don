const autoprefixer = require('autoprefixer');
const path = require('path');
const webpack = require('webpack');
const WebpackChunkHash = require('webpack-chunk-hash');

const byExtension = (...arr) =>
  arr.reduce(
    (r, obj) =>
      r.concat(
        Object.keys(obj).map(k =>
          Object.assign(
            {
              test: new RegExp(`\.(${k})$`),
            },
            obj[k]
          ))
      ),
    []
  );

const postcssLoader = {
  loader: 'postcss-loader',
  options: { plugins: () => [autoprefixer] },
};

module.exports = {
  devtool: 'source-map',
  output: {
    path: path.join(__dirname, '..', 'build'),
    filename: '[name]-bundle.js',
    publicPath: '/build/',
  },
  plugins: [
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify(
        process.env.NODE_ENV || 'development'
      ),
      'process.env.API_URL': JSON.stringify(process.env.API_URL || ''),
    }),
    process.env.NODE_ENV === 'production'
      ? new webpack.HashedModuleIdsPlugin()
      : new webpack.NamedModulesPlugin(),
    new WebpackChunkHash(),
  ],
  module: {
    rules: byExtension(
      {
        css: {
          include: /node_modules/,
          use: ['style-loader', 'css-loader?importLoaders=1', postcssLoader],
        },
      },
      {
        css: {
          exclude: /node_modules/,
          use: [
            'style-loader',
            process.env.NODE_ENV === 'production'
              ? 'css-loader?modules&importLoaders=1'
              : 'css-loader?modules&localIdentName=[path][name]---[local]---[hash:base64:5]&importLoaders=1',
            postcssLoader,
          ],
        },
        'woff2?(\\?v=[0-9]\\.[0-9]\\.[0-9])?': {
          use: ['url-loader?limit=10000&mimetype=application/font-woff'],
        },
        '(ttf|eot|svg)(\\?v=[0-9]\\.[0-9]\\.[0-9])?': {
          use: ['file-loader'],
        },
        js: {
          use: ['babel-loader'],
          include: path.join(__dirname, 'src'),
        },
        json: {
          use: ['json-loader'],
        },
        tpl: {
          use: ['template-string-loader'],
        },
        jpg: {
          use: [
            'url-loader?limit=10000&mimetype=image/jpeg',
            { loader: 'image-webpack-loader', query: {} },
          ],
        },
        png: {
          use: [
            'url-loader?limit=10000&mimetype=image/png',
            { loader: 'image-webpack-loader', query: {} },
          ],
        },
      }
    ),
  },
  resolve: {
    extensions: ['.js'],
    alias: {
      components: path.join(__dirname, 'src', 'components'),
      containers: path.join(__dirname, 'src', 'containers'),
      ducks: path.join(__dirname, 'src', 'ducks'),
      lib: path.join(__dirname, 'src', 'lib'),
      schema: path.join(__dirname, 'src', 'schema'),
    },
  },
};
