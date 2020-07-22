module.exports = {
  testRegex: ".*test\\.js$",
  coverageDirectory: './coverage',
  reporters: [ "default", "jest-junit" ],
  testURL: 'http://localhost',
  collectCoverageFrom: [
    '**/src/**/*.js',
    '!**/__tests__/**',
    '!**/node_modules/**',
  ],
  moduleNameMapper : {},
  verbose: true,
  automock: false,
  setupFiles: [
    "./setupJest.js",
  ],
  coverageThreshold: {
    global: {
      statements: 10,
      branches: 10,
      functions: 10,
      lines: 10,
    },
  },
  projects: ['./src'],
}
