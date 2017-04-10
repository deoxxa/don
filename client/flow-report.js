#!/usr/bin/env node

const { execFile } = require('child_process');
const { readFile, writeFile } = require('fs');
const { join } = require('path');
const { walk } = require('walk');

const files = [];

function sanitise(path) {
  return path.replace(/[\/\\\.]/g, '_') + '.html';
}

walk('src')
  .on('file', (rootPath, fileStats, next) => {
    if (!fileStats.name.endsWith('.js')) {
      return next();
    }

    const path = join(rootPath, fileStats.name);

    execFile('flow', ['coverage', '--json', path], (err, stdout) => {
      if (err) {
        return next(err);
      }

      const {
        expressions: {
          covered_count: covered = 0,
          uncovered_count: uncovered = 0,
          uncovered_locs: locations = [],
        },
      } = JSON.parse(stdout);

      locations.sort((a, b) => a.start.offset - b.start.offset);

      const total = covered + uncovered;

      let percentage = covered / total * 100;
      if (Number.isNaN(percentage)) {
        percentage = 100;
      }

      files.push({
        path,
        covered,
        uncovered,
        total,
        percentage,
      });

      readFile(path, { encoding: 'utf8' }, (err, fileContent) => {
        if (err) {
          return next(err);
        }

        let html = `
          <style>
            body { font-family: monospace; white-space: pre; }
            .u { color: red; background: #ffdddd; }
          </style>
        `;

        html += `<div><h1>${path}</h1>`;
        for (let i = 0; i < fileContent.length; i++) {
          locations.forEach(l => {
            if (l.start.offset === i) {
              html += `<span class="u" title="${l.start.line}:${l.start.column} - ${l.end.line}:${l.end.column}">`;
            }

            if (l.end.offset === i) {
              html += '</span>';
            }
          });

          switch (fileContent[i]) {
            case '<':
              html += '&lt;';
              break;
            case '>':
              html += '&gt;';
              break;
            case '&':
              html += '&amp;';
              break;
            case '\n':
              html += '<br>';
              break;
            default:
              html += fileContent[i];
              break;
          }
        }
        html += '</div>';

        writeFile(join('coverage', 'flow', sanitise(path)), html, next);
      });
    });
  })
  .on('end', () => {
    let covered = 0;
    let uncovered = 0;

    files.forEach(f => {
      covered += f.covered;
      uncovered += f.uncovered;
    });

    files.sort((a, b) => a.percentage - b.percentage);

    const menu = files.map(
      f =>
        `<li><a href="${sanitise(f.path)}" target="file">[${f.percentage.toFixed(2)}%] ${f.path}</a></li>`
    );

    const html = `
      <html>
        <head>
          <title>coverage</title>
          <style>
            #outer {
              display: flex;
              flex-direction: row;
              height: 100%;
            }
            #side {
              width: 250px;
              overflow-x: scroll;
            }
            #menu {
              list-style: none;
              margin: 0;
              padding: 0;
              white-space: nowrap;
            }
            #file {
              flex: 1;
              min-width: 1000px;
            }
          </style>
        </head>
        <body>
          <div id="outer">
            <div id="side">
              <h3>Overall: ${(covered / (uncovered + covered) * 100).toFixed(2)}%</h3>
              <ul id="menu">${menu.join('\n')}</ul>
            </div>
            <iframe id="file" name="file" src="${sanitise(files[0].path)}" />
          </div>
        </body>
      </html>
    `;

    writeFile(join('coverage', 'flow', 'index.html'), html, err => {
      if (err) {
        console.warn(err);
        process.exit(1);
      }
    });
  });
