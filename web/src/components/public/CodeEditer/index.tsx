import React, { useEffect, useState, useRef } from 'react'

import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/monokai.css'
import 'codemirror/keymap/sublime'
import 'codemirror/mode/shell/shell'
import { Controlled as CodeMirror } from 'react-codemirror2'
import styles from './index.less'

export default ({ code , onChange, readOnly= true, lineNumbers= false, height = 372 } : any ) => {
    const [statusCode, setStatusCode] = useState('')
    const codemirrorRef: any = useRef({});

    useEffect(()=> {
      setTimeout(()=> {
        setStatusCode(code)
      }, 100)
    }, [code])

    useEffect(()=> {
      // codemirrorRef?.current?.editor?.display.wrapper.style.height = height + "100px";
      if (codemirrorRef?.current) {
        const { editor } : any = codemirrorRef?.current || {}
        const { display = {}} = editor || {}
        const { wrapper = {}} = display || {}
        const { style = {}} = wrapper || {}
        style.height = height + 'px';
      }
    }, [codemirrorRef, height])

    return (
        <CodeMirror ref={codemirrorRef}
            value={ statusCode }
            className={ styles.code_wrapper}
            options={{
                theme: 'monokai',
                keyMap: 'sublime',
                mode : 'shell',
                lineWrapping: true,
                lineNumbers: lineNumbers,
                readOnly: readOnly,
            }}
            onBeforeChange={( editor: any , data: any , value: any ) => onChange( value )}
        />
    )
}