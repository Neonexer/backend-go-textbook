import React, { useState, useEffect } from "react";
import clsx from "clsx";
import styles from "./Quiz.module.css";

export interface Question {
  id: string;
  question: string;
  options: string[];
  correctIndex: number;
  explanation?: string;
}

interface QuizProps {
  quizId: string;
  title?: string;
  questions: Question[];
}

const LS = typeof window !== "undefined" ? window.localStorage : null;

function loadScore(quizId: string): number | null {
  if (!LS) return null;
  try {
    const saved = LS.getItem(`quiz-${quizId}`);
    return saved ? JSON.parse(saved).score : null;
  } catch {
    return null;
  }
}

function saveScore(quizId: string, score: number) {
  if (!LS) return;
  try {
    LS.setItem(`quiz-${quizId}`, JSON.stringify({ score, date: Date.now() }));
  } catch {
    // localStorage full or unavailable
  }
}

function clearScore(quizId: string) {
  if (!LS) return;
  try {
    LS.removeItem(`quiz-${quizId}`);
  } catch {
    // ignore
  }
}

export default function Quiz({ quizId, title, questions }: QuizProps) {
  const [selected, setSelected] = useState<Record<string, number>>({});
  const [submitted, setSubmitted] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [score, setScore] = useState<number | null>(null);

  useEffect(() => {
    setScore(loadScore(quizId));
    setMounted(true);
  }, [quizId]);

  useEffect(() => {
    if (mounted && score !== null) {
      saveScore(quizId, score);
    }
  }, [score, quizId, mounted]);

  const handleSelect = (questionId: string, optionIndex: number) => {
    if (submitted) return;
    setSelected((prev) => ({ ...prev, [questionId]: optionIndex }));
  };

  const handleSubmit = () => {
    const correct = questions.filter((q) => selected[q.id] === q.correctIndex).length;
    setScore(correct);
    setSubmitted(true);
  };

  const handleReset = () => {
    setSelected({});
    setSubmitted(false);
    setScore(null);
    clearScore(quizId);
  };

  const allAnswered = questions.every((q) => selected[q.id] !== undefined);

  return (
    <div className={styles.quiz}>
      <h3 className={styles.title}>{title || "Проверь себя"}</h3>

      {score !== null && (
        <div className={clsx(styles.scoreBanner, score === questions.length ? styles.perfect : styles.good)}>
          Результат: {score} / {questions.length}
          {score === questions.length ? " 🎉 Отлично!" : " — есть куда расти!"}
        </div>
      )}

      {questions.map((q) => {
        const isCorrect = submitted && selected[q.id] === q.correctIndex;
        const isWrong = submitted && selected[q.id] !== undefined && selected[q.id] !== q.correctIndex;

        return (
          <div key={q.id} className={styles.question}>
            <p className={styles.questionText}>{q.question}</p>
            <div className={styles.options}>
              {q.options.map((opt, i) => {
                const isSelected = selected[q.id] === i;
                const showCorrect = submitted && i === q.correctIndex;
                return (
                  <button
                    key={i}
                    className={clsx(
                      styles.option,
                      isSelected && styles.selected,
                      submitted && showCorrect && styles.correct,
                      submitted && isSelected && isWrong && styles.wrong
                    )}
                    onClick={() => handleSelect(q.id, i)}
                    disabled={submitted}
                  >
                    <span className={styles.optionLetter}>{String.fromCharCode(65 + i)}</span>
                    <span>{opt}</span>
                    {submitted && showCorrect && <span className={styles.checkmark}>✓</span>}
                    {submitted && isSelected && isWrong && <span className={styles.cross}>✗</span>}
                  </button>
                );
              })}
            </div>
            {submitted && isWrong && q.explanation && (
              <div className={styles.explanation}>{q.explanation}</div>
            )}
          </div>
        );
      })}

      <div className={styles.actions}>
        {!submitted ? (
          <button className={styles.submitBtn} onClick={handleSubmit} disabled={!allAnswered}>
            {allAnswered ? "Проверить" : `Ответь на все вопросы (${Object.keys(selected).length}/${questions.length})`}
          </button>
        ) : (
          <button className={styles.resetBtn} onClick={handleReset}>
            Пройти заново
          </button>
        )}
      </div>
    </div>
  );
}
