var tests = module.exports = {};

Object.defineProperties(tests, {
  blockTests: {
    get: require('require-all').bind(this, __dirname + '/BlockTests')
  },
  basicTests: {
    get: require('require-all').bind(this, __dirname + '/BasicTests/')
  },
  trieTests: {
    get: require('require-all').bind(this, __dirname + '/TrieTests/')
  },
  stateTests: {
    get: require('require-all').bind(this, __dirname + '/StateTests/')
  },
  vmTests: {
    get: require('require-all').bind(this, __dirname + '/VMTests')
  }
});
