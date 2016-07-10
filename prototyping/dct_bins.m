function [bins] = dct_bins(signal)
  n = rows(signal);
  transMat = zeros(n, n);
  for i = 1:n
    for j = 1:n
      transMat(i, j) = cos(pi / n * (j - 0.5) * (i - 1));
    end
  end
  bins = transMat*signal;
end
