import React from 'react'
import styled from 'styled-components'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/default.css'
import 'highlight.js/styles/github.css'

const Viewer = styled.div`
    background: #263238 !important;
    h1, h2, h3, h4, h5, h6 , p { margin : 0 ; }
    h1 { font-size:32px; }
    h2 { font-size:24px; }
    h3 { font-size:18.72px; }
    h4 { font-size:16px; }
    h5 { font-size:13.28px; }
    h6 { font-size:12px; }
    p , code , li , th , tr { color: #fff; } // #333
    p { white-space: nowrap; }
    ul { list-style-type: disc; }
    ol { list-style-type: decimal; }
    ul ul { list-style-type: circle; }
    ul ul ul { list-style-type: square; }
    table th, table tr , table td {
        border: 1px solid #dfe2e5;
        padding: 6px 13px;
    }
    table thead th { background-color: #F8F8F8; }
    table thead tr { background-color: #ffffff; }
    table tr:nth-child(2n) { background-color: #f8f8f8; }
    code, pre { background-color: #263238; } // #f9f9f9
    pre { padding : 8px }
    //
    .hljs-meta .hljs-string, .hljs-regexp, .hljs-string {
        color: #005cc5
    }
`

marked.setOptions({
    renderer: new marked.Renderer(),
    highlight: function (code: string) {
        return hljs.highlightAuto(code).value;
    },
    pedantic: false,
    gfm: true,
    breaks: false,
    sanitize: false,
    smartLists: true,
    langPrefix: "language-",
    smartypants: false,
    xhtml: false
});

export default (props: any) => {
    const { code } = props
    if (!code) return <span></span>
    return (
        <Viewer
            className="hljs"
            dangerouslySetInnerHTML={{ __html: marked.parse(code) }}
        />
    )
}