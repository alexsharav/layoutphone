export function buildPreviewHtml(html: string, css: string, js: string) {
  return `<!doctype html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <style>
      * { box-sizing: border-box; }
      body { margin: 0; font-family: Inter, Arial, sans-serif; }
      ${css}
    </style>
  </head>
  <body>
    ${html}
    <script>${js}<\/script>
  </body>
</html>`;
}
