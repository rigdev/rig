import React, { useRef, useState, useEffect } from "react"
import { CSSTransition, SwitchTransition } from "react-transition-group"

import useIsBrowser from "@docusaurus/useIsBrowser"
import { useLocation } from "@docusaurus/router"
import uuid from "react-uuid"

import Button from "../Button"
import { useUser } from "../../providers/User"

const Feedback = ({
  event,
  question = "Was this page helpful?",
  positiveBtn = "Yes",
  negativeBtn = "No",
  positiveQuestion = "What was most helpful?",
  negativeQuestion = "What can we improve?",
  submitBtn = "Submit",
  submitMessage = "Thank you for helping improve our documentation!",
  showPossibleSolutions = true,
  className = "",
}) => {
  const [showForm, setShowForm] = useState(false)
  const [submittedFeedback, setSubmittedFeedback] = useState(false)
  const [loading, setLoading] = useState(false)
  const inlineFeedbackRef = useRef(null)
  const inlineQuestionRef = useRef(null)
  const inlineMessageRef = useRef(null)
  const [positiveFeedback, setPositiveFeedback] = useState(false)
  const [message, setMessage] = useState("")
  const [id, setId] = useState(null)
  const nodeRef = submittedFeedback
    ? inlineMessageRef
    : showForm
    ? inlineQuestionRef
    : inlineFeedbackRef

  const isBrowser = useIsBrowser()
  const location = useLocation()
  const { track } = useUser()

  function handleFeedback(e) {
    const feedback = e.target.classList.contains("positive")
    setPositiveFeedback(feedback)
    setShowForm(true)
    submitFeedback(e, feedback)
  }

  function submitFeedback(e, feedback = null) {
    if (isBrowser) {
      if (showForm) {
        setLoading(true)
      }
      track(
        event,
        {
          url: location.pathname,
          label: document.title,
          feedback:
            (feedback !== null && feedback) ||
            (feedback === null && positiveFeedback)
              ? "yes"
              : "no",
          message: message?.length ? message : null,
        },
        function () {
          if (showForm) {
            setLoading(false)
            resetForm()
          }
        }
      )
    }
  }

  function resetForm() {
    setShowForm(false)
    setSubmittedFeedback(true)
    if (message) {
      setId(null)
    }
  }

  useEffect(() => {
    if (!id) {
      setId(uuid())
    }
  }, [id])

  return (
    <div className={`${className}`}>          
        {submittedFeedback && (
            <>
                <hr class="solid" />
                <h3>
                    {question}
                </h3>
                <p>
                    {submitMessage}
                </p>
                <hr class="solid" />
            </>
        )}
        {!submittedFeedback && (<>
        <hr class="solid" />
        {!showForm && !submittedFeedback && (
            <div>
                <h3>
                    {question}
                </h3>
                <div
                    ref={inlineFeedbackRef}
                    style={{display: "flex"}}
                >
                <Button
                    onClick={handleFeedback}
                >
                    {positiveBtn}
                </Button>
                <div style={{marginRight: "10px"}}/>
                <Button
                    onClick={handleFeedback}
                >
                    {negativeBtn}
                </Button>
                </div>
            </div>
        )}
        {showForm && !submittedFeedback && (
            <div ref={inlineQuestionRef}>
                <div>
                    <p>
                        {positiveFeedback ? positiveQuestion : negativeQuestion}
                    </p>
                    <textarea
                        rows={4}
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                    ></textarea>
                </div>
                <Button
                    onClick={submitFeedback}
                    disabled={loading}
                    className="tw-mt-1 tw-w-fit"
                >
                    {submitBtn}
                </Button>
            </div>
        )}
        <hr class="solid" />
        </>)}
    </div>
  )
}

export default Feedback