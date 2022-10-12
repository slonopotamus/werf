package stage

type Image interface {
  GetBuildContext() *build_context.BuildContext
}
