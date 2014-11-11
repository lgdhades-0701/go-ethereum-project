#!/usr/bin/env node

'use strict';

var path = require('path');

var del = require('del');
var gulp = require('gulp');
var browserify = require('browserify');
var jshint = require('gulp-jshint');
var uglify = require('gulp-uglify');
var rename = require('gulp-rename');
var envify = require('envify/custom');
var unreach = require('unreachable-branch-transform');
var source = require('vinyl-source-stream');
var exorcist = require('exorcist');
var bower = require('bower');

var DEST = './dist/';

var build = function(src, dst) {
  return browserify({
      debug: true,
      insert_global_vars: false,
      detectGlobals: false,
      bundleExternal: false
    })
    .add('./')
    .require('./' + src + '.js', {expose: 'web3'})
    .transform('envify', {
      NODE_ENV: 'build'
    })
    .transform('unreachable-branch-transform')
    .transform('uglifyify', {
      mangle: false,
      compress: {
        dead_code: false,
        conditionals: true,
        unused: false,
        hoist_funs: true,
        hoist_vars: true,
        negate_iife: false
      },
      beautify: true,
      warnings: true
    })
    .bundle()
    .pipe(exorcist(path.join( DEST, dst + '.js.map')))
    .pipe(source(dst + '.js'))
    .pipe(gulp.dest( DEST ));
};

var uglifyFile = function(file) {
  return gulp.src( DEST + file + '.js')
    .pipe(uglify())
    .pipe(rename(file + '.min.js'))
    .pipe(gulp.dest( DEST ));
};

gulp.task('bower', function(cb){
  bower.commands.install().on('end', function (installed){
    console.log(installed);
    cb();
  });
});

gulp.task('lint', function(){
  return gulp.src(['./*.js', './lib/*.js'])
    .pipe(jshint())
    .pipe(jshint.reporter('default'));
});

gulp.task('clean', ['lint'], function(cb) {
  del([ DEST ], cb);
});

gulp.task('build', ['clean'], function () {
    return build('index', 'ethereum');
});

gulp.task('buildQt', ['clean'], function () {
    return build('index_qt', 'ethereum_qt');
});

gulp.task('uglify', ['build'], function(){
    return uglifyFile('ethereum');
});

gulp.task('uglifyQt', ['buildQt'], function () {
    return uglifyFile('ethereum_qt');
});

gulp.task('watch', function() {
  gulp.watch(['./lib/*.js'], ['lint', 'prepare', 'build']);
});

gulp.task('default', ['bower', 'lint', 'build', 'uglify']);
gulp.task('qt', ['bower', 'lint', 'buildQt', 'uglifyQt']);
