# spotr templates

Program templates are JSON files that can be imported into spotr.
You can import a whole program template, or import one workout from a template
into the active program.

By default, spotr reads templates from `templates/programs/`. Set
`SPOTR_TEMPLATE_DIR` to use a local template directory outside the repo.

## Add a Template

Anyone can add a community template by opening a PR with a JSON file under
`templates/programs/`.

1. Create or export a template:

   ```text
   :template export My Program templates/programs/my-program.json
   ```

2. Review the generated JSON.
3. Validate the template:

   ```text
   :template validate templates/programs/my-program.json
   ```

4. Run the test suite:

   ```bash
   go test ./...
   ```

5. Commit the JSON file and open a PR. CI runs the same test suite on PRs,
   and the PR template includes a template contribution checklist.

Each file should:

- Use the lowercase dash-separated template name as the filename, like
  `Push Pull Legs` -> `push-pull-legs.json`.
- Match `templates/schema/program-template.schema.json`.
- Include at least one workout.
- Include at least one exercise per workout.
- Use non-negative integer `sets` and `reps`.

## Format

Templates use this shape:

```json
{
  "name": "Push Pull Legs",
  "description": "Six-day hypertrophy split",
  "version": 1,
  "workouts": [
    {
      "name": "Push",
      "exercises": [
        { "name": "Bench Press", "sets": 3, "reps": 8 }
      ]
    }
  ]
}
```
