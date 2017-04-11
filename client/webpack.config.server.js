const UnusedFilesWebpackPlugin = require('unused-files-webpack-plugin').UnusedFilesWebpackPlugin;

module.exports = require('./webpack.config.common');

module.exports.entry = {
  'entry-server': ['./src/entry-server'],
};

module.exports.output.libraryTarget = 'commonjs2';

if (process.env.NODE_ENV === 'production') {
  module.exports.plugins.push(
    new UnusedFilesWebpackPlugin({
      pattern: 'src/**/*.*',
      globOptions: {
        ignore: ['src/**/__tests__/*.js', 'src/entry-*.js'],
      },
      failOnUnused: true,
    })
  );
}

module.exports.module.rules.forEach(rule => {
  if (rule.use.indexOf('style-loader') === 0) {
    rule.use.shift();
    rule.use[0] = rule.use[0].replace('css-loader', 'css-loader/locals');
  }
});
