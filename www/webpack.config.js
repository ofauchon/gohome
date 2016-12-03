var path = require('path');
var webpack = require('webpack');

var minify = false;

module.exports = {
    entry: './assets/js/gohome.js',
    output: {
        path: './assets/js/',
        filename: 'gohome-out.js'
    },
    module: {
        loaders: [
            {
                test: /\.jsx?$/,
                loader: 'babel-loader',
                exclude: /node_modules/,
                query: {
                    cacheDirectory: true,
                    presets: ['es2015', 'react']
                }
            },
            {
                test: /\.less$/,
                loader: 'style-loader!css-loader!less-loader'
            },
        ],
        postLoaders: [
            {
                test: /\.js$/,
                exclude: /node_modules/, // do not lint third-party code
                loader: 'jshint-loader'
            }
        ]
    },
    jshint: {
        strict: "global"
    },
    plugins: minify ? [
        new webpack.optimize.UglifyJsPlugin({
            minimize: true
        })
    ] : []
};
